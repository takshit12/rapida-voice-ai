from google.protobuf import timestamp_pb2 as _timestamp_pb2
import app.bridges.artifacts.protos.common_pb2 as _common_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AssistantTool(_message.Message):
    __slots__ = ("id", "assistantId", "name", "description", "fields", "executionMethod", "executionOptions", "status", "createdDate", "updatedDate")
    ID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    FIELDS_FIELD_NUMBER: _ClassVar[int]
    EXECUTIONMETHOD_FIELD_NUMBER: _ClassVar[int]
    EXECUTIONOPTIONS_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    CREATEDDATE_FIELD_NUMBER: _ClassVar[int]
    UPDATEDDATE_FIELD_NUMBER: _ClassVar[int]
    id: int
    assistantId: int
    name: str
    description: str
    fields: _struct_pb2.Struct
    executionMethod: str
    executionOptions: _containers.RepeatedCompositeFieldContainer[_common_pb2.Metadata]
    status: str
    createdDate: _timestamp_pb2.Timestamp
    updatedDate: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[int] = ..., assistantId: _Optional[int] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., fields: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., executionMethod: _Optional[str] = ..., executionOptions: _Optional[_Iterable[_Union[_common_pb2.Metadata, _Mapping]]] = ..., status: _Optional[str] = ..., createdDate: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updatedDate: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class CreateAssistantToolRequest(_message.Message):
    __slots__ = ("assistantId", "name", "description", "fields", "executionMethod", "executionOptions")
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    FIELDS_FIELD_NUMBER: _ClassVar[int]
    EXECUTIONMETHOD_FIELD_NUMBER: _ClassVar[int]
    EXECUTIONOPTIONS_FIELD_NUMBER: _ClassVar[int]
    assistantId: int
    name: str
    description: str
    fields: _struct_pb2.Struct
    executionMethod: str
    executionOptions: _containers.RepeatedCompositeFieldContainer[_common_pb2.Metadata]
    def __init__(self, assistantId: _Optional[int] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., fields: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., executionMethod: _Optional[str] = ..., executionOptions: _Optional[_Iterable[_Union[_common_pb2.Metadata, _Mapping]]] = ...) -> None: ...

class UpdateAssistantToolRequest(_message.Message):
    __slots__ = ("id", "assistantId", "name", "description", "fields", "executionMethod", "executionOptions")
    ID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    FIELDS_FIELD_NUMBER: _ClassVar[int]
    EXECUTIONMETHOD_FIELD_NUMBER: _ClassVar[int]
    EXECUTIONOPTIONS_FIELD_NUMBER: _ClassVar[int]
    id: int
    assistantId: int
    name: str
    description: str
    fields: _struct_pb2.Struct
    executionMethod: str
    executionOptions: _containers.RepeatedCompositeFieldContainer[_common_pb2.Metadata]
    def __init__(self, id: _Optional[int] = ..., assistantId: _Optional[int] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., fields: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., executionMethod: _Optional[str] = ..., executionOptions: _Optional[_Iterable[_Union[_common_pb2.Metadata, _Mapping]]] = ...) -> None: ...

class GetAssistantToolRequest(_message.Message):
    __slots__ = ("id", "assistantId")
    ID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    id: int
    assistantId: int
    def __init__(self, id: _Optional[int] = ..., assistantId: _Optional[int] = ...) -> None: ...

class DeleteAssistantToolRequest(_message.Message):
    __slots__ = ("id", "assistantId")
    ID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    id: int
    assistantId: int
    def __init__(self, id: _Optional[int] = ..., assistantId: _Optional[int] = ...) -> None: ...

class GetAssistantToolResponse(_message.Message):
    __slots__ = ("code", "success", "data", "error")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: AssistantTool
    error: _common_pb2.Error
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Union[AssistantTool, _Mapping]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ...) -> None: ...

class GetAllAssistantToolRequest(_message.Message):
    __slots__ = ("assistantId", "paginate", "criterias")
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    PAGINATE_FIELD_NUMBER: _ClassVar[int]
    CRITERIAS_FIELD_NUMBER: _ClassVar[int]
    assistantId: int
    paginate: _common_pb2.Paginate
    criterias: _containers.RepeatedCompositeFieldContainer[_common_pb2.Criteria]
    def __init__(self, assistantId: _Optional[int] = ..., paginate: _Optional[_Union[_common_pb2.Paginate, _Mapping]] = ..., criterias: _Optional[_Iterable[_Union[_common_pb2.Criteria, _Mapping]]] = ...) -> None: ...

class GetAllAssistantToolResponse(_message.Message):
    __slots__ = ("code", "success", "data", "error", "paginated")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    PAGINATED_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: _containers.RepeatedCompositeFieldContainer[AssistantTool]
    error: _common_pb2.Error
    paginated: _common_pb2.Paginated
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Iterable[_Union[AssistantTool, _Mapping]]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ..., paginated: _Optional[_Union[_common_pb2.Paginated, _Mapping]] = ...) -> None: ...

class GetAllAssistantToolLogRequest(_message.Message):
    __slots__ = ("projectId", "paginate", "criterias", "order")
    PROJECTID_FIELD_NUMBER: _ClassVar[int]
    PAGINATE_FIELD_NUMBER: _ClassVar[int]
    CRITERIAS_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    projectId: int
    paginate: _common_pb2.Paginate
    criterias: _containers.RepeatedCompositeFieldContainer[_common_pb2.Criteria]
    order: _common_pb2.Ordering
    def __init__(self, projectId: _Optional[int] = ..., paginate: _Optional[_Union[_common_pb2.Paginate, _Mapping]] = ..., criterias: _Optional[_Iterable[_Union[_common_pb2.Criteria, _Mapping]]] = ..., order: _Optional[_Union[_common_pb2.Ordering, _Mapping]] = ...) -> None: ...

class GetAssistantToolLogRequest(_message.Message):
    __slots__ = ("projectId", "id")
    PROJECTID_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    projectId: int
    id: int
    def __init__(self, projectId: _Optional[int] = ..., id: _Optional[int] = ...) -> None: ...

class GetAssistantToolLogResponse(_message.Message):
    __slots__ = ("code", "success", "data", "error")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: AssistantToolLog
    error: _common_pb2.Error
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Union[AssistantToolLog, _Mapping]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ...) -> None: ...

class GetAllAssistantToolLogResponse(_message.Message):
    __slots__ = ("code", "success", "data", "error", "paginated")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    PAGINATED_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: _containers.RepeatedCompositeFieldContainer[AssistantToolLog]
    error: _common_pb2.Error
    paginated: _common_pb2.Paginated
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Iterable[_Union[AssistantToolLog, _Mapping]]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ..., paginated: _Optional[_Union[_common_pb2.Paginated, _Mapping]] = ...) -> None: ...

class AssistantToolLog(_message.Message):
    __slots__ = ("id", "action", "request", "response", "status", "createdDate", "updatedDate", "assistantId", "projectId", "organizationId", "assistantConversationId", "assistantConversationMessageId", "assetPrefix", "executionMethod", "timeTaken", "assistantToolId", "assistantToolName", "assistantTool")
    ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    CREATEDDATE_FIELD_NUMBER: _ClassVar[int]
    UPDATEDDATE_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    PROJECTID_FIELD_NUMBER: _ClassVar[int]
    ORGANIZATIONID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTCONVERSATIONID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTCONVERSATIONMESSAGEID_FIELD_NUMBER: _ClassVar[int]
    ASSETPREFIX_FIELD_NUMBER: _ClassVar[int]
    EXECUTIONMETHOD_FIELD_NUMBER: _ClassVar[int]
    TIMETAKEN_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTTOOLID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTTOOLNAME_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTTOOL_FIELD_NUMBER: _ClassVar[int]
    id: int
    action: _struct_pb2.Struct
    request: _struct_pb2.Struct
    response: _struct_pb2.Struct
    status: str
    createdDate: _timestamp_pb2.Timestamp
    updatedDate: _timestamp_pb2.Timestamp
    assistantId: int
    projectId: int
    organizationId: int
    assistantConversationId: int
    assistantConversationMessageId: str
    assetPrefix: str
    executionMethod: str
    timeTaken: int
    assistantToolId: int
    assistantToolName: str
    assistantTool: AssistantTool
    def __init__(self, id: _Optional[int] = ..., action: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., request: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., response: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., status: _Optional[str] = ..., createdDate: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updatedDate: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., assistantId: _Optional[int] = ..., projectId: _Optional[int] = ..., organizationId: _Optional[int] = ..., assistantConversationId: _Optional[int] = ..., assistantConversationMessageId: _Optional[str] = ..., assetPrefix: _Optional[str] = ..., executionMethod: _Optional[str] = ..., timeTaken: _Optional[int] = ..., assistantToolId: _Optional[int] = ..., assistantToolName: _Optional[str] = ..., assistantTool: _Optional[_Union[AssistantTool, _Mapping]] = ...) -> None: ...
