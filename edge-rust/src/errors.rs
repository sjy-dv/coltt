// src/errors.rs
use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("Collection already exists")]
    UniqueViolation,

    #[error("Collection not found")]
    NotFound,

    #[error("Dimension mismatch")]
    DimensionMismatch,

    #[error("Serialization error: {0}")]
    SerializationError(String),

    #[error("Other error: {0}")]
    Other(String),
}
