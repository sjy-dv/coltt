// src/id_generator.rs
use std::sync::atomic::{AtomicU64, Ordering};
use lazy_static::lazy_static;

/// 전역 ID 생성기
pub struct IdGenerator {
    counter: AtomicU64,
}

impl IdGenerator {
    pub fn new(start: u64) -> Self {
        Self {
            counter: AtomicU64::new(start),
        }
    }

    /// 다음 고유 ID를 반환합니다.
    pub fn next_id(&self) -> u64 {
        self.counter.fetch_add(1, Ordering::Relaxed)
    }
}

lazy_static! {
    /// 전역적으로 사용되는 ID 생성기
    pub static ref ID_GENERATOR: IdGenerator = IdGenerator::new(1); // 시작 ID는 1
}
