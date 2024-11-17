from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class StorageType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    highspeed_memory: _ClassVar[StorageType]
    stable_disk: _ClassVar[StorageType]

class Distance(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    L2sq: _ClassVar[Distance]
    Ip: _ClassVar[Distance]
    Cosine: _ClassVar[Distance]
    Haversine: _ClassVar[Distance]
    Divergence: _ClassVar[Distance]
    Pearson: _ClassVar[Distance]
    Hamming: _ClassVar[Distance]
    Tanimoto: _ClassVar[Distance]
    Sorensen: _ClassVar[Distance]

class Quantization(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    None: _ClassVar[Quantization]
    BF16: _ClassVar[Quantization]
    F16: _ClassVar[Quantization]
    F32: _ClassVar[Quantization]
    F64: _ClassVar[Quantization]
    I8: _ClassVar[Quantization]
    B1: _ClassVar[Quantization]

class ErrorCode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    UNDEFINED: _ClassVar[ErrorCode]
    RPC_ERROR: _ClassVar[ErrorCode]
    COMMUNICATION_SHARD_RPC_ERROR: _ClassVar[ErrorCode]
    COMMUNICATION_SHARD_ERROR: _ClassVar[ErrorCode]
    MARSHAL_ERROR: _ClassVar[ErrorCode]
    INTERNAL_FUNC_ERROR: _ClassVar[ErrorCode]
highspeed_memory: StorageType
stable_disk: StorageType
L2sq: Distance
Ip: Distance
Cosine: Distance
Haversine: Distance
Divergence: Distance
Pearson: Distance
Hamming: Distance
Tanimoto: Distance
Sorensen: Distance
None: Quantization
BF16: Quantization
F16: Quantization
F32: Quantization
F64: Quantization
I8: Quantization
B1: Quantization
UNDEFINED: ErrorCode
RPC_ERROR: ErrorCode
COMMUNICATION_SHARD_RPC_ERROR: ErrorCode
COMMUNICATION_SHARD_ERROR: ErrorCode
MARSHAL_ERROR: ErrorCode
INTERNAL_FUNC_ERROR: ErrorCode

class Collection(_message.Message):
    __slots__ = ("collection_name", "distance", "quantization", "dim", "connectivity", "expansion_add", "expansion_search", "multi", "storage")
    COLLECTION_NAME_FIELD_NUMBER: _ClassVar[int]
    DISTANCE_FIELD_NUMBER: _ClassVar[int]
    QUANTIZATION_FIELD_NUMBER: _ClassVar[int]
    DIM_FIELD_NUMBER: _ClassVar[int]
    CONNECTIVITY_FIELD_NUMBER: _ClassVar[int]
    EXPANSION_ADD_FIELD_NUMBER: _ClassVar[int]
    EXPANSION_SEARCH_FIELD_NUMBER: _ClassVar[int]
    MULTI_FIELD_NUMBER: _ClassVar[int]
    STORAGE_FIELD_NUMBER: _ClassVar[int]
    collection_name: str
    distance: Distance
    quantization: Quantization
    dim: int
    connectivity: int
    expansion_add: int
    expansion_search: int
    multi: bool
    storage: StorageType
    def __init__(self, collection_name: _Optional[str] = ..., distance: _Optional[_Union[Distance, str]] = ..., quantization: _Optional[_Union[Quantization, str]] = ..., dim: _Optional[int] = ..., connectivity: _Optional[int] = ..., expansion_add: _Optional[int] = ..., expansion_search: _Optional[int] = ..., multi: bool = ..., storage: _Optional[_Union[StorageType, str]] = ...) -> None: ...

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

class CollectionList(_message.Message):
    __slots__ = ("collection", "collection_size", "collection_memory")
    COLLECTION_FIELD_NUMBER: _ClassVar[int]
    COLLECTION_SIZE_FIELD_NUMBER: _ClassVar[int]
    COLLECTION_MEMORY_FIELD_NUMBER: _ClassVar[int]
    collection: Collection
    collection_size: int
    collection_memory: int
    def __init__(self, collection: _Optional[_Union[Collection, _Mapping]] = ..., collection_size: _Optional[int] = ..., collection_memory: _Optional[int] = ...) -> None: ...

class CollectionLists(_message.Message):
    __slots__ = ("collections", "status", "error")
    COLLECTIONS_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    collections: _containers.RepeatedCompositeFieldContainer[CollectionList]
    status: bool
    error: Error
    def __init__(self, collections: _Optional[_Iterable[_Union[CollectionList, _Mapping]]] = ..., status: bool = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class DeleteCollectionResponse(_message.Message):
    __slots__ = ("status", "error")
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    status: bool
    error: Error
    def __init__(self, status: bool = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class CollectionName(_message.Message):
    __slots__ = ("collection_name", "with_size")
    COLLECTION_NAME_FIELD_NUMBER: _ClassVar[int]
    WITH_SIZE_FIELD_NUMBER: _ClassVar[int]
    collection_name: str
    with_size: bool
    def __init__(self, collection_name: _Optional[str] = ..., with_size: bool = ...) -> None: ...

class GetCollections(_message.Message):
    __slots__ = ("with_size",)
    WITH_SIZE_FIELD_NUMBER: _ClassVar[int]
    with_size: bool
    def __init__(self, with_size: bool = ...) -> None: ...

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

class SystemInfo(_message.Message):
    __slots__ = ("uptime", "cpu_load1", "cpu_load5", "cpu_load15", "mem_total", "mem_available", "mem_used", "mem_free", "mem_used_percent")
    UPTIME_FIELD_NUMBER: _ClassVar[int]
    CPU_LOAD1_FIELD_NUMBER: _ClassVar[int]
    CPU_LOAD5_FIELD_NUMBER: _ClassVar[int]
    CPU_LOAD15_FIELD_NUMBER: _ClassVar[int]
    MEM_TOTAL_FIELD_NUMBER: _ClassVar[int]
    MEM_AVAILABLE_FIELD_NUMBER: _ClassVar[int]
    MEM_USED_FIELD_NUMBER: _ClassVar[int]
    MEM_FREE_FIELD_NUMBER: _ClassVar[int]
    MEM_USED_PERCENT_FIELD_NUMBER: _ClassVar[int]
    uptime: int
    cpu_load1: float
    cpu_load5: float
    cpu_load15: float
    mem_total: int
    mem_available: int
    mem_used: int
    mem_free: int
    mem_used_percent: float
    def __init__(self, uptime: _Optional[int] = ..., cpu_load1: _Optional[float] = ..., cpu_load5: _Optional[float] = ..., cpu_load15: _Optional[float] = ..., mem_total: _Optional[int] = ..., mem_available: _Optional[int] = ..., mem_used: _Optional[int] = ..., mem_free: _Optional[int] = ..., mem_used_percent: _Optional[float] = ...) -> None: ...
