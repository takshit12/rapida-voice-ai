from google.protobuf import struct_pb2 as _struct_pb2
import app.bridges.artifacts.protos.common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GeneralConnectRequest(_message.Message):
    __slots__ = ("state", "code", "scope", "connect")
    STATE_FIELD_NUMBER: _ClassVar[int]
    CODE_FIELD_NUMBER: _ClassVar[int]
    SCOPE_FIELD_NUMBER: _ClassVar[int]
    CONNECT_FIELD_NUMBER: _ClassVar[int]
    state: str
    code: str
    scope: str
    connect: str
    def __init__(self, state: _Optional[str] = ..., code: _Optional[str] = ..., scope: _Optional[str] = ..., connect: _Optional[str] = ...) -> None: ...

class GeneralConnectResponse(_message.Message):
    __slots__ = ("code", "success", "provider", "redirectTo", "error")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    REDIRECTTO_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    provider: str
    redirectTo: str
    error: _common_pb2.Error
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., provider: _Optional[str] = ..., redirectTo: _Optional[str] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ...) -> None: ...

class GetConnectorFilesRequest(_message.Message):
    __slots__ = ("paginate", "criterias", "provider")
    PAGINATE_FIELD_NUMBER: _ClassVar[int]
    CRITERIAS_FIELD_NUMBER: _ClassVar[int]
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    paginate: _common_pb2.Paginate
    criterias: _containers.RepeatedCompositeFieldContainer[_common_pb2.Criteria]
    provider: str
    def __init__(self, paginate: _Optional[_Union[_common_pb2.Paginate, _Mapping]] = ..., criterias: _Optional[_Iterable[_Union[_common_pb2.Criteria, _Mapping]]] = ..., provider: _Optional[str] = ...) -> None: ...

class GetConnectorFilesResponse(_message.Message):
    __slots__ = ("code", "success", "data", "paginated", "error", "args")
    class ArgsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    PAGINATED_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    ARGS_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: _containers.RepeatedCompositeFieldContainer[_struct_pb2.Struct]
    paginated: _common_pb2.Paginated
    error: _common_pb2.Error
    args: _containers.ScalarMap[str, str]
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Iterable[_Union[_struct_pb2.Struct, _Mapping]]] = ..., paginated: _Optional[_Union[_common_pb2.Paginated, _Mapping]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ..., args: _Optional[_Mapping[str, str]] = ...) -> None: ...
