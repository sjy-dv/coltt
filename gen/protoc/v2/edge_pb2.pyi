from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Distance(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    Cosine: _ClassVar[Distance]
    Euclidean: _ClassVar[Distance]

class Quantization(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    None: _ClassVar[Quantization]
    F16: _ClassVar[Quantization]

class ErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNDEFINED: _ClassVar[ErrorCode]
    RPC_ERROR: _ClassVar[ErrorCode]
    COMMUNICATION_SHARD_RPC_ERROR: _ClassVar[ErrorCode]
    COMMUNICATION_SHARD_ERROR: _ClassVar[ErrorCode]
    MARSHAL_ERROR: _ClassVar[ErrorCode]
    INTERNAL_FUNC_ERROR: _ClassVar[ErrorCode]
Cosine: Distance
Euclidean: Distance
None: Quantization
F16: Quantization
UNDEFINED: ErrorCode
RPC_ERROR: ErrorCode
COMMUNICATION_SHARD_RPC_ERROR: ErrorCode
COMMUNICATION_SHARD_ERROR: ErrorCode
MARSHAL_ERROR: ErrorCode
INTERNAL_FUNC_ERROR: ErrorCode

class Collection(_message.Message):
    __slots__ = ("collection_name", "distance", "quantization", "dim")
    COLLECTION_NAME_FIELD_NUMBER: _ClassVar[int]
    DISTANCE_FIELD_NUMBER: _ClassVar[int]
    QUANTIZATION_FIELD_NUMBER: _ClassVar[int]
    DIM_FIELD_NUMBER: _ClassVar[int]
    collection_name: str
    distance: Distance
    quantization: Quantization
    dim: int
    def __init__(self, collection_name: _Optional[str] = ..., distance: _Optional[_Union[Distance, str]] = ..., quantization: _Optional[_Union[Quantization, str]] = ..., dim: _Optional[int] = ...) -> None: ...

class CollectionResponse(_message.Message):
    __slots__ = ("collection", "status", "error")
    COLLECTION_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    collection: Collection
    status: bool
    error: Error
    def __init__(self, collection: _Optional[_Union[Collection, _Mapping]] = ..., status: bool = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class CollectionDetail(_message.Message):
    __slots__ = ("collection", "collection_size", "collection_memory", "status", "error")
    COLLECTION_FIELD_NUMBER: _ClassVar[int]
    COLLECTION_SIZE_FIELD_NUMBER: _ClassVar[int]
    COLLECTION_MEMORY_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    collection: Collection
    collection_size: int
    collection_memory: int
    status: bool
    error: Error
    def __init__(self, collection: _Optional[_Union[Collection, _Mapping]] = ..., collection_size: _Optional[int] = ..., collection_memory: _Optional[int] = ..., status: bool = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class DeleteCollectionResponse(_message.Message):
    __slots__ = ("status", "error")
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    status: bool
    error: Error
    def __init__(self, status: bool = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class ModifyDataset(_message.Message):
    __slots__ = ("id", "collection_name", "vector", "metadata")
    ID_FIELD_NUMBER: _ClassVar[int]
    COLLECTION_NAME_FIELD_NUMBER: _ClassVar[int]
    VECTOR_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    id: str
    collection_name: str
    vector: _containers.RepeatedScalarFieldContainer[float]
    metadata: _struct_pb2.Struct
    def __init__(self, id: _Optional[str] = ..., collection_name: _Optional[str] = ..., vector: _Optional[_Iterable[float]] = ..., metadata: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class CollectionName(_message.Message):
    __slots__ = ("collection_name", "with_size")
    COLLECTION_NAME_FIELD_NUMBER: _ClassVar[int]
    WITH_SIZE_FIELD_NUMBER: _ClassVar[int]
    collection_name: str
    with_size: bool
    def __init__(self, collection_name: _Optional[str] = ..., with_size: bool = ...) -> None: ...

class DeleteDataset(_message.Message):
    __slots__ = ("id", "collection_name")
    ID_FIELD_NUMBER: _ClassVar[int]
    COLLECTION_NAME_FIELD_NUMBER: _ClassVar[int]
    id: str
    collection_name: str
    def __init__(self, id: _Optional[str] = ..., collection_name: _Optional[str] = ...) -> None: ...

class Response(_message.Message):
    __slots__ = ("status", "error")
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    status: bool
    error: Error
    def __init__(self, status: bool = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class Error(_message.Message):
    __slots__ = ("error_message", "error_code")
    ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    ERROR_CODE_FIELD_NUMBER: _ClassVar[int]
    error_message: str
    error_code: ErrorCode
    def __init__(self, error_message: _Optional[str] = ..., error_code: _Optional[_Union[ErrorCode, str]] = ...) -> None: ...

class SearchReq(_message.Message):
    __slots__ = ("collection_name", "vector", "topK", "filter", "with_latency")
    class FilterEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    COLLECTION_NAME_FIELD_NUMBER: _ClassVar[int]
    VECTOR_FIELD_NUMBER: _ClassVar[int]
    TOPK_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    WITH_LATENCY_FIELD_NUMBER: _ClassVar[int]
    collection_name: str
    vector: _containers.RepeatedScalarFieldContainer[float]
    topK: int
    filter: _containers.ScalarMap[str, str]
    with_latency: bool
    def __init__(self, collection_name: _Optional[str] = ..., vector: _Optional[_Iterable[float]] = ..., topK: _Optional[int] = ..., filter: _Optional[_Mapping[str, str]] = ..., with_latency: bool = ...) -> None: ...

class SearchResponse(_message.Message):
    __slots__ = ("status", "error", "candidates", "latency")
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    CANDIDATES_FIELD_NUMBER: _ClassVar[int]
    LATENCY_FIELD_NUMBER: _ClassVar[int]
    status: bool
    error: Error
    candidates: _containers.RepeatedCompositeFieldContainer[Candidates]
    latency: str
    def __init__(self, status: bool = ..., error: _Optional[_Union[Error, _Mapping]] = ..., candidates: _Optional[_Iterable[_Union[Candidates, _Mapping]]] = ..., latency: _Optional[str] = ...) -> None: ...

class Candidates(_message.Message):
    __slots__ = ("id", "metadata", "score")
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    SCORE_FIELD_NUMBER: _ClassVar[int]
    id: str
    metadata: _struct_pb2.Struct
    score: float
    def __init__(self, id: _Optional[str] = ..., metadata: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., score: _Optional[float] = ...) -> None: ...
