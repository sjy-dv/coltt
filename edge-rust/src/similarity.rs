// src/similarity.rs
use crate::db::Distance;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, Serialize, Deserialize, JsonSchema, PartialEq, Eq)]
pub enum DistanceMetric {
    #[serde(rename = "euclidean")]
    Euclidean,
    #[serde(rename = "cosine")]
    Cosine,
    #[serde(rename = "dot")]
    DotProduct,
}

pub fn get_cache_attr(metric: Distance, vec: &[f32]) -> f32 {
    match metric {
        // Dot product과 Euclidean은 캐싱을 허용하지 않음
        Distance::DotProduct | Distance::Euclidean => 0.0,
        // Cosine은 벡터의 크기를 미리 계산
        Distance::Cosine => vec.iter().map(|&x| x.powi(2)).sum::<f32>().sqrt(),
    }
}

pub fn get_distance_fn(metric: Distance) -> impl Fn(&[f32], &[f32], f32) -> f32 {
    match metric {
        Distance::Euclidean => euclidean_distance,
        // Cosine과 DotProduct는 동일한 함수 사용 (Cosine의 경우 정규화된 벡터를 사용)
        Distance::Cosine | Distance::DotProduct => dot_product,
    }
}

fn euclidean_distance(a: &[f32], b: &[f32], a_sum_squares: f32) -> f32 {
    let mut cross_terms = 0.0;
    let mut b_sum_squares = 0.0;

    for (i, j) in a.iter().zip(b) {
        cross_terms += i * j;
        b_sum_squares += j.powi(2);
    }

    2.0f32
        .mul_add(-cross_terms, a_sum_squares + b_sum_squares)
        .max(0.0)
        .sqrt()
}

fn dot_product(a: &[f32], b: &[f32], _: f32) -> f32 {
    a.iter().zip(b).fold(0.0, |acc, (x, y)| acc + x * y)
}

pub fn normalize(vec: &[f32]) -> Vec<f32> {
    let magnitude = (vec.iter().fold(0.0, |acc, &val| acc + val * val)).sqrt();

    if magnitude > std::f32::EPSILON {
        vec.iter().map(|&val| val / magnitude).collect()
    } else {
        vec.to_vec()
    }
}
