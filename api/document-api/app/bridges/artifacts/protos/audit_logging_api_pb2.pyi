from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
import app.bridges.artifacts.protos.common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AuditLog(_message.Message):
    __slots__ = ("id", "integrationName", "assetPrefix", "responseStatus", "timeTaken", "status", "projectId", "organizationId", "credentialId", "externalAuditMetadatas", "createdDate", "updatedDate", "request", "response", "metrics")
    ID_FIELD_NUMBER: _ClassVar[int]
    INTEGRATIONNAME_FIELD_NUMBER: _ClassVar[int]
    ASSETPREFIX_FIELD_NUMBER: _ClassVar[int]
    RESPONSESTATUS_FIELD_NUMBER: _ClassVar[int]
    TIMETAKEN_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    PROJECTID_FIELD_NUMBER: _ClassVar[int]
    ORGANIZATIONID_FIELD_NUMBER: _ClassVar[int]
    CREDENTIALID_FIELD_NUMBER: _ClassVar[int]
    EXTERNALAUDITMETADATAS_FIELD_NUMBER: _ClassVar[int]
    CREATEDDATE_FIELD_NUMBER: _ClassVar[int]
    UPDATEDDATE_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    METRICS_FIELD_NUMBER: _ClassVar[int]
    id: int
    integrationName: str
    assetPrefix: str
    responseStatus: int
    timeTaken: int
    status: str
    projectId: int
    organizationId: int
    credentialId: int
    externalAuditMetadatas: _containers.RepeatedCompositeFieldContainer[_common_pb2.Metadata]
    createdDate: _timestamp_pb2.Timestamp
    updatedDate: _timestamp_pb2.Timestamp
    request: _struct_pb2.Struct
    response: _struct_pb2.Struct
    metrics: _containers.RepeatedCompositeFieldContainer[_common_pb2.Metric]
    def __init__(self, id: _Optional[int] = ..., integrationName: _Optional[str] = ..., assetPrefix: _Optional[str] = ..., responseStatus: _Optional[int] = ..., timeTaken: _Optional[int] = ..., status: _Optional[str] = ..., projectId: _Optional[int] = ..., organizationId: _Optional[int] = ..., credentialId: _Optional[int] = ..., externalAuditMetadatas: _Optional[_Iterable[_Union[_common_pb2.Metadata, _Mapping]]] = ..., createdDate: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updatedDate: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., request: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., response: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., metrics: _Optional[_Iterable[_Union[_common_pb2.Metric, _Mapping]]] = ...) -> None: ...

class GetAllAuditLogRequest(_message.Message):
    __slots__ = ("paginate", "criterias", "projectId", "organizationId")
    PAGINATE_FIELD_NUMBER: _ClassVar[int]
    CRITERIAS_FIELD_NUMBER: _ClassVar[int]
    PROJECTID_FIELD_NUMBER: _ClassVar[int]
    ORGANIZATIONID_FIELD_NUMBER: _ClassVar[int]
    paginate: _common_pb2.Paginate
    criterias: _containers.RepeatedCompositeFieldContainer[_common_pb2.Criteria]
    projectId: int
    organizationId: int
    def __init__(self, paginate: _Optional[_Union[_common_pb2.Paginate, _Mapping]] = ..., criterias: _Optional[_Iterable[_Union[_common_pb2.Criteria, _Mapping]]] = ..., projectId: _Optional[int] = ..., organizationId: _Optional[int] = ...) -> None: ...

class GetAllAuditLogResponse(_message.Message):
    __slots__ = ("code", "success", "data", "error", "paginated")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    PAGINATED_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: _containers.RepeatedCompositeFieldContainer[AuditLog]
    error: _common_pb2.Error
    paginated: _common_pb2.Paginated
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Iterable[_Union[AuditLog, _Mapping]]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ..., paginated: _Optional[_Union[_common_pb2.Paginated, _Mapping]] = ...) -> None: ...

class GetAuditLogRequest(_message.Message):
    __slots__ = ("id", "projectId", "organizationId")
    ID_FIELD_NUMBER: _ClassVar[int]
    PROJECTID_FIELD_NUMBER: _ClassVar[int]
    ORGANIZATIONID_FIELD_NUMBER: _ClassVar[int]
    id: int
    projectId: int
    organizationId: int
    def __init__(self, id: _Optional[int] = ..., projectId: _Optional[int] = ..., organizationId: _Optional[int] = ...) -> None: ...

class GetAuditLogResponse(_message.Message):
    __slots__ = ("code", "success", "data", "error")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: AuditLog
    error: _common_pb2.Error
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Union[AuditLog, _Mapping]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ...) -> None: ...

class CreateMetadataRequest(_message.Message):
    __slots__ = ("id", "additionalData")
    class AdditionalDataEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ID_FIELD_NUMBER: _ClassVar[int]
    ADDITIONALDATA_FIELD_NUMBER: _ClassVar[int]
    id: int
    additionalData: _containers.ScalarMap[str, str]
    def __init__(self, id: _Optional[int] = ..., additionalData: _Optional[_Mapping[str, str]] = ...) -> None: ...

class CreateMetadataResponse(_message.Message):
    __slots__ = ("code", "success", "data", "error")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: AuditLog
    error: _common_pb2.Error
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Union[AuditLog, _Mapping]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ...) -> None: ...
