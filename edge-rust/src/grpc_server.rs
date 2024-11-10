use edge::edge_server::Edge;
// src/grpc_server.rs
use tonic::{Request, Response, Status};
use crate::db::{Db, Embedding, SimilarityResult, Distance, DbError};
use crate::similarity::{get_distance_fn, normalize};
use std::sync::Arc;
use tokio::sync::RwLock;
use serde_json::{Map, Value as JsonValue};
use prost_types::{Struct, Value as ProtoValue};
use std::collections::HashMap;
use bitvec::prelude::*;
use ordered_float::OrderedFloat;

pub mod edge {
    tonic::include_proto!("edge");
}

use edge::{
    CreateCollectionRequest, CreateCollectionResponse, DeleteCollectionRequest, DeleteCollectionResponse,
    GetCollectionRequest, GetCollectionResponse, AddEmbeddingRequest, AddEmbeddingResponse,
    UpdateEmbeddingRequest, UpdateEmbeddingResponse, RemoveEmbeddingRequest, RemoveEmbeddingResponse,
    SearchVectorsRequest, SearchVectorsResponse, SearchFiltersRequest, SearchFiltersResponse,
    SearchHybridRequest, SearchHybridResponse, Embedding as ProtoEmbedding, Collection as ProtoCollection,
    SimilarityResult as ProtoSimilarityResult, Distance as ProtoDistance,
};

#[derive(Debug)]
pub struct MyVectorDB {
    db: Arc<RwLock<Db>>,
}

impl MyVectorDB {
    pub fn new(db: Arc<RwLock<Db>>) -> Self {
        Self { db }
    }
}

#[tonic::async_trait]
impl Edge for MyVectorDB { // VectorDb → VectorDB로 수정
    // Collection Management

    async fn create_collection(
        &self,
        request: Request<CreateCollectionRequest>,
    ) -> Result<Response<CreateCollectionResponse>, Status> {
        let req = request.into_inner();
        let mut db = self.db.write().await;

        let distance = match ProtoDistance::try_from(req.distance) {
            Ok(d) => match d {
                ProtoDistance::Euclidean => Distance::Euclidean,
                ProtoDistance::Cosine => Distance::Cosine,
                ProtoDistance::DotProduct => Distance::DotProduct,
            },
            Err(_) => return Err(Status::invalid_argument("Invalid distance metric")),
        };

        match db.create_collection(req.name, req.dimension as usize, distance, req.metadata_fields) {
            Ok(_) => Ok(Response::new(CreateCollectionResponse { success: true })),
            Err(DbError::UniqueViolation) => Err(Status::already_exists("Collection already exists")),
            Err(_) => Err(Status::unknown("Unknown error occurred")),
        }
    }

    // DeleteCollection
    async fn delete_collection(
        &self,
        request: Request<DeleteCollectionRequest>,
    ) -> Result<Response<DeleteCollectionResponse>, Status> {
        let req = request.into_inner();
        let mut db = self.db.write().await;

        match db.delete_collection(&req.name) {
            Ok(_) => Ok(Response::new(DeleteCollectionResponse { success: true })),
            Err(DbError::NotFound) => Err(Status::not_found("Collection not found")),
            Err(_) => Err(Status::unknown("Unknown error occurred")),
        }
    }

    // GetCollection
    async fn get_collection(
        &self,
        request: Request<GetCollectionRequest>,
    ) -> Result<Response<GetCollectionResponse>, Status> {
        let req = request.into_inner();
        let db = self.db.read().await;

        match db.get_collection(&req.name) {
            Some(collection) => {
                let proto_distance = match collection.distance {
                    Distance::Euclidean => ProtoDistance::Euclidean,
                    Distance::Cosine => ProtoDistance::Cosine,
                    Distance::DotProduct => ProtoDistance::DotProduct,
                };

                let proto_collection = ProtoCollection {
                    name: collection.name.clone(),
                    dimension: collection.dimension as u32,
                    distance: proto_distance as i32,
                };

                Ok(Response::new(GetCollectionResponse {
                    collection: Some(proto_collection),
                }))
            }
            None => Err(Status::not_found("Collection not found")),
        }
    }

    // AddEmbedding
    async fn add_embedding(
        &self,
        request: Request<AddEmbeddingRequest>,
    ) -> Result<Response<AddEmbeddingResponse>, Status> {
        let req = request.into_inner();
        let mut db = self.db.write().await;

        // embedding 필드가 Some인지 확인
        let embedding_proto = req.embedding.ok_or(Status::invalid_argument("Embedding is required"))?;

        // Convert protobuf Struct to HashMap<String, JsonValue>
        let metadata = if let Some(metadata_struct) = embedding_proto.metadata {
            let metadata_map = metadata_struct
                .fields
                .into_iter()
                .map(|(k, v)| (k, convert_proto_value(v)))
                .collect::<HashMap<String, JsonValue>>();
            Some(metadata_map)
        } else {
            None
        };

        let embedding = Embedding {
            id: 0, // 서버에서 할당
            vector: embedding_proto.vector,
            metadata,
            bitmap: BitVec::new(), // 초기화는 서버에서 수행
        };

        match db.add_embedding(&req.collection_name, embedding) {
            Ok(id) => Ok(Response::new(AddEmbeddingResponse { id })),
            Err(DbError::NotFound) => Err(Status::not_found("Collection not found")),
            Err(DbError::DimensionMismatch) => Err(Status::invalid_argument("Dimension mismatch")),
            Err(_) => Err(Status::unknown("Failed to add embedding")),
        }
    }

    // UpdateEmbedding
    async fn update_embedding(
        &self,
        request: Request<UpdateEmbeddingRequest>,
    ) -> Result<Response<UpdateEmbeddingResponse>, Status> {
        let req = request.into_inner();
        let mut db = self.db.write().await;

        // embedding 필드가 Some인지 확인
        let updated_embedding_proto = req.embedding.ok_or(Status::invalid_argument("Embedding is required"))?;

        // Convert protobuf Struct to HashMap<String, JsonValue>
        let metadata = if let Some(metadata_struct) = updated_embedding_proto.metadata {
            let metadata_map = metadata_struct
                .fields
                .into_iter()
                .map(|(k, v)| (k, convert_proto_value(v)))
                .collect::<HashMap<String, JsonValue>>();
            Some(metadata_map)
        } else {
            None
        };

        let updated_embedding = Embedding {
            id: updated_embedding_proto.id,
            vector: updated_embedding_proto.vector,
            metadata,
            bitmap: BitVec::new(), // 서버에서 비트맵 생성
        };

        match db.update_embedding(&req.collection_name, updated_embedding) {
            Ok(_) => Ok(Response::new(UpdateEmbeddingResponse { success: true })),
            Err(DbError::NotFound) => Err(Status::not_found("Embedding not found")),
            Err(DbError::DimensionMismatch) => Err(Status::invalid_argument("Dimension mismatch")),
            Err(_) => Err(Status::unknown("Failed to update embedding")),
        }
    }

    // RemoveEmbedding
    async fn remove_embedding(
        &self,
        request: Request<RemoveEmbeddingRequest>,
    ) -> Result<Response<RemoveEmbeddingResponse>, Status> {
        let req = request.into_inner();
        let mut db = self.db.write().await;

        match db.remove_embedding(&req.collection_name, req.id) {
            Ok(_) => Ok(Response::new(RemoveEmbeddingResponse { success: true })),
            Err(DbError::NotFound) => Err(Status::not_found("Embedding not found")),
            Err(_) => Err(Status::unknown("Failed to remove embedding")),
        }
    }

    // SearchVectors
    async fn search_vectors(
        &self,
        request: Request<SearchVectorsRequest>,
    ) -> Result<Response<SearchVectorsResponse>, Status> {
        let req = request.into_inner();
        let db = self.db.read().await;

        let similarity_results = db
            .get_similarity_vectors_only(&req.collection_name, &req.query, req.k as usize)
            .map_err(|_| Status::invalid_argument("Failed to perform vector search"))?;

        let proto_results = similarity_results
            .into_iter()
            .map(|res| {
                ProtoSimilarityResult {
                    score: res.score.0, // OrderedFloat에서 f32 추출
                    embedding: Some(convert_embedding_to_proto(&res.embedding)),
                }
            })
            .collect();

        Ok(Response::new(SearchVectorsResponse { results: proto_results }))
    }

    // SearchFilters
    async fn search_filters(
        &self,
        request: Request<SearchFiltersRequest>,
    ) -> Result<Response<SearchFiltersResponse>, Status> {
        let req = request.into_inner();
        let db = self.db.read().await;

        let similarity_results = db
            .get_similarity_filters_only(
                &req.collection_name,
                Some(convert_filters_struct(req.filters)),
                req.k as usize,
            )
            .map_err(|_| Status::invalid_argument("Failed to perform filter search"))?;

        let proto_results = similarity_results
            .into_iter()
            .map(|res| {
                ProtoSimilarityResult {
                    score: res.score.0, // OrderedFloat에서 f32 추출
                    embedding: Some(convert_embedding_to_proto(&res.embedding)),
                }
            })
            .collect();

        Ok(Response::new(SearchFiltersResponse { results: proto_results }))
    }

    // SearchHybrid
    async fn search_hybrid(
        &self,
        request: Request<SearchHybridRequest>,
    ) -> Result<Response<SearchHybridResponse>, Status> {
        let req = request.into_inner();
        let db = self.db.read().await;

        let similarity_results = db
            .get_similarity_hybrid(
                &req.collection_name,
                &req.query,
                Some(convert_filters_struct(req.filters)),
                req.k as usize,
            )
            .map_err(|_| Status::invalid_argument("Failed to perform hybrid search"))?;

        let proto_results = similarity_results
            .into_iter()
            .map(|res| {
                ProtoSimilarityResult {
                    score: res.score.0, // OrderedFloat에서 f32 추출
                    embedding: Some(convert_embedding_to_proto(&res.embedding)),
                }
            })
            .collect();

        Ok(Response::new(SearchHybridResponse { results: proto_results }))
    }
}

// Helper functions

use prost_types::value::Kind as ProtoKind;

/// Protobuf `Value`를 Serde JSON `Value`로 변환
fn convert_proto_value(value: ProtoValue) -> JsonValue {
    match value.kind {
        Some(ProtoKind::NullValue(_)) => JsonValue::Null,
        Some(ProtoKind::NumberValue(n)) => JsonValue::from(n),
        Some(ProtoKind::StringValue(s)) => JsonValue::from(s),
        Some(ProtoKind::BoolValue(b)) => JsonValue::from(b),
        Some(ProtoKind::StructValue(s)) => convert_struct_to_json(s),
        Some(ProtoKind::ListValue(l)) => JsonValue::from(
            l.values
                .into_iter()
                .map(convert_proto_value)
                .collect::<Vec<_>>(),
        ),
        None => JsonValue::Null,
    }
}

/// Protobuf `Struct`를 Serde JSON `Value::Object`로 변환
fn convert_struct_to_json(s: Struct) -> JsonValue {
    let map: Map<String, JsonValue> = s
        .fields
        .into_iter()
        .map(|(k, v)| (k, convert_proto_value(v)))
        .collect();
    JsonValue::Object(map)
}

/// 필터 `Struct`를 Serde JSON `Map`으로 변환
fn convert_filters_struct(filters_struct: Option<Struct>) -> HashMap<String, JsonValue> {
    if let Some(s) = filters_struct {
        s.fields
            .into_iter()
            .map(|(k, v)| (k, convert_proto_value(v)))
            .collect()
    } else {
        HashMap::new()
    }
}

/// Serde JSON `Value`를 Protobuf `Value`로 변환
fn convert_json_value(value: JsonValue) -> ProtoValue {
    match value {
        JsonValue::Null => ProtoValue {
            kind: Some(ProtoKind::NullValue(prost_types::NullValue::NullValue as i32)),
        },
        JsonValue::Bool(b) => ProtoValue {
            kind: Some(ProtoKind::BoolValue(b)),
        },
        JsonValue::Number(n) => {
            if let Some(i) = n.as_i64() {
                ProtoValue {
                    kind: Some(ProtoKind::NumberValue(i as f64)),
                }
            } else if let Some(u) = n.as_u64() {
                ProtoValue {
                    kind: Some(ProtoKind::NumberValue(u as f64)),
                }
            } else if let Some(f) = n.as_f64() {
                ProtoValue {
                    kind: Some(ProtoKind::NumberValue(f)),
                }
            } else {
                ProtoValue {
                    kind: Some(ProtoKind::NullValue(prost_types::NullValue::NullValue as i32)),
                }
            }
        }
        JsonValue::String(s) => ProtoValue {
            kind: Some(ProtoKind::StringValue(s)),
        },
        JsonValue::Array(arr) => ProtoValue {
            kind: Some(ProtoKind::ListValue(prost_types::ListValue {
                values: arr
                    .into_iter()
                    .map(convert_json_value)
                    .collect::<Vec<_>>(),
            })),
        },
        JsonValue::Object(obj) => {
            let fields = obj
                .into_iter()
                .map(|(k, v)| (k, convert_json_value(v)))
                .collect();
            ProtoValue {
                kind: Some(ProtoKind::StructValue(Struct { fields })),
            }
        }
    }
}

/// `Embedding`을 Protobuf `Embedding`으로 변환
fn convert_embedding_to_proto(embedding: &Embedding) -> ProtoEmbedding {
    let proto_metadata = if let Some(metadata_map) = &embedding.metadata {
        let fields = metadata_map
            .iter()
            .map(|(k, v)| (k.clone(), convert_json_value(v.clone())))
            .collect();
        Some(Struct { fields })
    } else {
        None
    };

    ProtoEmbedding {
        id: embedding.id,
        vector: embedding.vector.clone(),
        metadata: proto_metadata,
    }
}
