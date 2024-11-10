// src/main.rs
#![warn(clippy::all, clippy::pedantic, clippy::nursery)]

use anyhow::Result;
use std::sync::Arc;

mod db;
mod errors;
mod grpc_server;
mod similarity;
mod id_generator;

use grpc_server::MyVectorDB;
use tonic::transport::Server as TonicServer;

pub mod edge {
    tonic::include_proto!("edge"); // "edge"는 Protobuf 파일의 package 이름과 일치해야 합니다.
}

#[tokio::main]
async fn main() -> Result<()> {


    // 데이터베이스 로드 또는 초기화
    let db = Arc::new(tokio::sync::RwLock::new(db::Db::load_from_store()?));

    // gRPC 서버를 실행
    let grpc_db = db.clone();
    let grpc_handle = tokio::spawn(async move {
        let addr = "[::1]:50051".parse().unwrap();
        let vector_db = MyVectorDB::new(grpc_db);

        tracing::info!("gRPC Server listening on {}", addr);

        TonicServer::builder()
            .add_service(edge::edge_server::EdgeServer::new(vector_db)) // 수정된 부분
            .serve(addr)
            .await
            .unwrap();
    });

    // gRPC 서버가 완료될 때까지 대기
    grpc_handle.await?; // `?` 연산자 한 번만 사용

    Ok(())
}
