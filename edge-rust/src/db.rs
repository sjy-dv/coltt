// src/db.rs
use crate::id_generator::ID_GENERATOR;
use serde::{Deserialize, Serialize};
use serde_json::{Map, Value as JsonValue};
use bitvec::prelude::*;
use std::collections::HashMap;
use rayon::prelude::*;
use anyhow::Context;
use std::fs;
use std::path::Path;
use thiserror::Error;
use bincode; // 이진 직렬화
use crate::similarity::{normalize, get_cache_attr, get_distance_fn}; // 필요한 함수 불러오기
// src/db.rs
use ordered_float::OrderedFloat;

// ... (기존 코드)


#[derive(Debug, Clone, PartialEq)]
pub struct SimilarityResult {
    pub score: OrderedFloat<f32>, // 변경: f32 -> OrderedFloat<f32>
    pub embedding: Embedding,
}

impl PartialOrd for SimilarityResult {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        self.score.partial_cmp(&other.score)
    }
}

impl Ord for SimilarityResult {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        self.score.cmp(&other.score)
    }
}

impl Eq for SimilarityResult {}

const STORE_PATH: &str = "store.db";


#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Embedding {
    pub id: u64, // 고유 ID
    pub vector: Vec<f32>,
    pub metadata: Option<HashMap<String, JsonValue>>, // 동적 메타데이터
    #[serde(skip)]
    pub bitmap: BitVec, // 서버 내부에서 사용
}



#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Collection {
    pub name: String,
    pub dimension: usize,
    pub distance: Distance,
    #[serde(default)]
    pub embeddings: Vec<Embedding>,
    pub metadata_fields: Vec<String>,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum Distance {
    Euclidean,
    Cosine,
    DotProduct,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Db {
    pub collections: HashMap<String, Collection>,
}

#[derive(Debug, Error)]
pub enum DbError {
    #[error("Collection already exists")]
    UniqueViolation,

    #[error("Collection not found")]
    NotFound,

    #[error("Dimension mismatch")]
    DimensionMismatch,
}

impl Db {
    /// 새로운 데이터베이스 인스턴스를 생성
    pub fn new() -> Self {
        Self {
            collections: HashMap::new(),
        }
    }

    /// 저장된 데이터베이스를 로드
    pub fn load_from_store() -> anyhow::Result<Self> {
        if !Path::new(STORE_PATH).exists() {
            tracing::debug!("Store file not found. Creating new database.");
            return Ok(Self::new());
        }

        tracing::debug!("Loading database from store.");
        let data = fs::read(STORE_PATH).context("Failed to read store file")?;
        let db = bincode::deserialize(&data).context("Failed to deserialize database")?;
        Ok(db)
    }

    /// 데이터베이스 상태를 저장
    pub fn save_to_store(&self) -> anyhow::Result<()> {
        let data = bincode::serialize(self).context("Failed to serialize database")?;
        fs::write(STORE_PATH, data).context("Failed to write to store file")?;
        Ok(())
    }

    /// 컬렉션 생성
    pub fn create_collection(
        &mut self,
        name: String,
        dimension: usize,
        distance: Distance,
        metadata_fields: Vec<String>,
    ) -> Result<Collection, DbError> {
        if self.collections.contains_key(&name) {
            return Err(DbError::UniqueViolation);
        }

        let collection = Collection {
            name: name.clone(),
            dimension,
            distance,
            embeddings: Vec::new(),
            metadata_fields,
        };

        self.collections.insert(name, collection.clone());

        Ok(collection)
    }

    /// 컬렉션 삭제
    pub fn delete_collection(&mut self, name: &str) -> Result<(), DbError> {
        if self.collections.remove(name).is_some() {
            Ok(())
        } else {
            Err(DbError::NotFound)
        }
    }

    /// 컬렉션 조회
    pub fn get_collection(&self, name: &str) -> Option<&Collection> {
        self.collections.get(name)
    }

    /// Embedding 추가
    pub fn add_embedding(
        &mut self,
        collection_name: &str,
        mut embedding: Embedding,
    ) -> Result<u64, DbError> {
        let collection = self
            .collections
            .get_mut(collection_name)
            .ok_or(DbError::NotFound)?;

        if embedding.vector.len() != collection.dimension {
            return Err(DbError::DimensionMismatch);
        }

        // 새로운 고유 ID 할당
        embedding.id = ID_GENERATOR.next_id();

        // 메타데이터 기반으로 비트맵 생성
        embedding.bitmap =
            Self::generate_bitmap_from_metadata(&embedding.metadata, &collection.metadata_fields);

        // 벡터 정규화 (코사인 거리의 경우)
        if collection.distance == Distance::Cosine {
            embedding.vector = normalize(&embedding.vector);
        }

        collection.embeddings.push(embedding.clone());

        Ok(embedding.id)
    }

    /// Embedding 업데이트
    pub fn update_embedding(
        &mut self,
        collection_name: &str,
        updated_embedding: Embedding,
    ) -> Result<(), DbError> {
        let collection = self
            .collections
            .get_mut(collection_name)
            .ok_or(DbError::NotFound)?;

        let pos = collection.embeddings.iter().position(|e| e.id == updated_embedding.id)
            .ok_or(DbError::NotFound)?;

        if updated_embedding.vector.len() != collection.dimension {
            return Err(DbError::DimensionMismatch);
        }

        // 메타데이터 기반으로 비트맵 생성
        let bitmap = Self::generate_bitmap_from_metadata(&updated_embedding.metadata, &collection.metadata_fields);

        // 업데이트된 Embedding 적용
        collection.embeddings[pos].vector = updated_embedding.vector;
        collection.embeddings[pos].metadata = updated_embedding.metadata;
        collection.embeddings[pos].bitmap = bitmap;

        // 벡터 정규화 (코사인 거리의 경우)
        if collection.distance == Distance::Cosine {
            collection.embeddings[pos].vector = normalize(&collection.embeddings[pos].vector);
        }

        Ok(())
    }

    /// Embedding 삭제
    pub fn remove_embedding(&mut self, collection_name: &str, id: u64) -> Result<(), DbError> {
        let collection = self
            .collections
            .get_mut(collection_name)
            .ok_or(DbError::NotFound)?;

        let pos = collection.embeddings.iter().position(|e| e.id == id)
            .ok_or(DbError::NotFound)?;

        collection.embeddings.remove(pos);
        Ok(())
    }

    /// 벡터만 검색
    pub fn get_similarity_vectors_only(
        &self,
        collection_name: &str,
        query: &[f32],
        k: usize,
    ) -> Result<Vec<SimilarityResult>, DbError> {
        let collection = self
            .collections
            .get(collection_name)
            .ok_or(DbError::NotFound)?;

        let similarity_results = collection.get_similarity(query, k);

        Ok(similarity_results)
    }

    /// 필터만 검색
    pub fn get_similarity_filters_only(
        &self,
        collection_name: &str,
        filters: Option<HashMap<String, JsonValue>>,
        k: usize,
    ) -> Result<Vec<SimilarityResult>, DbError> {
        let collection = self
            .collections
            .get(collection_name)
            .ok_or(DbError::NotFound)?;

        let filter_bitmap = Self::generate_bitmap_from_filters(filters, &collection.metadata_fields);

        let similarity_results = collection.get_similarity_filters(filter_bitmap, k);

        Ok(similarity_results)
    }

    /// 하이브리드 검색 (벡터 + 필터)
    pub fn get_similarity_hybrid(
        &self,
        collection_name: &str,
        query: &[f32],
        filters: Option<HashMap<String, JsonValue>>,
        k: usize,
    ) -> Result<Vec<SimilarityResult>, DbError> {
        let collection = self
            .collections
            .get(collection_name)
            .ok_or(DbError::NotFound)?;

        let filter_bitmap = Self::generate_bitmap_from_filters(filters, &collection.metadata_fields);

        let similarity_results = collection.get_similarity_hybrid(query, filter_bitmap, k);

        Ok(similarity_results)
    }

    /// 메타데이터 기반으로 비트맵 생성
    fn generate_bitmap_from_metadata(
        metadata: &Option<HashMap<String, JsonValue>>,
        metadata_fields: &Vec<String>,
    ) -> BitVec {
        let mut bitmap = BitVec::repeat(false, metadata_fields.len());
        if let Some(meta) = metadata {
            for (i, field) in metadata_fields.iter().enumerate() {
                if meta.contains_key(field) {
                    bitmap.set(i, true);
                }
            }
        }
        bitmap
    }

    /// 필터 기반으로 비트맵 생성
    fn generate_bitmap_from_filters(
        filters: Option<HashMap<String, JsonValue>>,
        metadata_fields: &Vec<String>,
    ) -> BitVec {
        let mut bitmap = BitVec::repeat(false, metadata_fields.len());
        if let Some(filter_map) = filters {
            for (i, field) in metadata_fields.iter().enumerate() {
                if filter_map.contains_key(field) {
                    bitmap.set(i, true);
                }
            }
        }
        bitmap
    }
}

impl Drop for Db {
    fn drop(&mut self) {
        if let Err(e) = self.save_to_store() {
            tracing::error!("Failed to save database on drop: {:?}", e);
        }
    }
}
impl Collection {
    // ... (기존 코드)

    /// 벡터만 기반으로 유사도 검색
    pub fn get_similarity(&self, query: &[f32], k: usize) -> Vec<SimilarityResult> {
        let memo_attr = get_cache_attr(self.distance, query);
        let distance_fn = get_distance_fn(self.distance);

        let scores = self
            .embeddings
            .par_iter()
            .map(|embedding| {
                let score = distance_fn(&embedding.vector, query, memo_attr);
                SimilarityResult {
                    score: OrderedFloat(score),
                    embedding: embedding.clone(),
                }
            })
            .collect::<Vec<_>>();

        let mut heap: std::collections::BinaryHeap<SimilarityResult> = std::collections::BinaryHeap::new();
        for res in scores {
            if heap.len() < k || res.score < heap.peek().unwrap().score {
                heap.push(res);
                if heap.len() > k {
                    heap.pop();
                }
            }
        }

        heap.into_sorted_vec()
    }

    /// 필터만 기반으로 유사도 검색
    pub fn get_similarity_filters(&self, filter_bitmap: BitVec, k: usize) -> Vec<SimilarityResult> {
        let distance_fn = get_distance_fn(self.distance);

        let scores = self
            .embeddings
            .par_iter()
            .filter(|embedding| {
                // 필터 비트맵과 embedding의 비트맵 비교
                bitmap_matches(&embedding.bitmap, &filter_bitmap)
            })
            .map(|embedding| {
                // 필터만 검색 시 score는 0.0으로 설정
                SimilarityResult {
                    score: OrderedFloat(0.0),
                    embedding: embedding.clone(),
                }
            })
            .collect::<Vec<_>>();

        // Top-k 정렬 (필터만 검색 시 정렬 기준이 없으므로 임의로 정렬)
        scores.into_iter().take(k).collect()
    }

    /// 벡터와 필터를 모두 기반으로 유사도 검색
    pub fn get_similarity_hybrid(&self, query: &[f32], filter_bitmap: BitVec, k: usize) -> Vec<SimilarityResult> {
        let memo_attr = get_cache_attr(self.distance, query);
        let distance_fn = get_distance_fn(self.distance);

        let scores = self
            .embeddings
            .par_iter()
            .filter(|embedding| {
                // 필터 비트맵과 embedding의 비트맵 비교
                bitmap_matches(&embedding.bitmap, &filter_bitmap)
            })
            .map(|embedding| {
                let score = distance_fn(&embedding.vector, query, memo_attr);
                SimilarityResult {
                    score: OrderedFloat(score),
                    embedding: embedding.clone(),
                }
            })
            .collect::<Vec<_>>();

        let mut heap: std::collections::BinaryHeap<SimilarityResult> = std::collections::BinaryHeap::new();
        for res in scores {
            if heap.len() < k || res.score < heap.peek().unwrap().score {
                heap.push(res);
                if heap.len() > k {
                    heap.pop();
                }
            }
        }

        heap.into_sorted_vec()
    }
}

/// Helper function to check if `filter_bitmap` is a subset of `embedding_bitmap`
fn bitmap_matches(embedding_bitmap: &BitVec, filter_bitmap: &BitVec) -> bool {
    let emb_slice = embedding_bitmap.as_bitslice();
    let filt_slice = filter_bitmap.as_bitslice();

    // Ensure both slices have the same length
    if emb_slice.len() != filt_slice.len() {
        return false;
    }

    emb_slice.iter().zip(filt_slice.iter()).all(|(e, f)| !(*f) || *e)
}