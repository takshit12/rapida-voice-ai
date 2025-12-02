import app.bridges.artifacts.protos.common_pb2 as _common_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ConnectAssistantAction(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetAssistantAction(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ExecuteAssistantAction(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class AssistantActionRequest(_message.Message):
    __slots__ = ("connect", "get", "execute")
    CONNECT_FIELD_NUMBER: _ClassVar[int]
    GET_FIELD_NUMBER: _ClassVar[int]
    EXECUTE_FIELD_NUMBER: _ClassVar[int]
    connect: ConnectAssistantAction
    get: GetAssistantAction
    execute: ExecuteAssistantAction
    def __init__(self, connect: _Optional[_Union[ConnectAssistantAction, _Mapping]] = ..., get: _Optional[_Union[GetAssistantAction, _Mapping]] = ..., execute: _Optional[_Union[ExecuteAssistantAction, _Mapping]] = ...) -> None: ...

class AssistantActionResult(_message.Message):
    __slots__ = ("success", "data")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    success: bool
    data: str
    def __init__(self, success: bool = ..., data: _Optional[str] = ...) -> None: ...

class AssistantActionResponse(_message.Message):
    __slots__ = ("code", "success", "error", "result")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    error: _common_pb2.Error
    result: AssistantActionResult
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ..., result: _Optional[_Union[AssistantActionResult, _Mapping]] = ...) -> None: ...

class AgentTalkRequest(_message.Message):
    __slots__ = ("configuration", "message")
    CONFIGURATION_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    configuration: _common_pb2.AssistantConversationConfiguration
    message: _common_pb2.AssistantConversationUserMessage
    def __init__(self, configuration: _Optional[_Union[_common_pb2.AssistantConversationConfiguration, _Mapping]] = ..., message: _Optional[_Union[_common_pb2.AssistantConversationUserMessage, _Mapping]] = ...) -> None: ...

class AgentTalkResponse(_message.Message):
    __slots__ = ("code", "success", "error", "interruption", "assistant", "disconnectAction", "holdAction", "assistantTransferAction", "phoneCallTransferAction")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTION_FIELD_NUMBER: _ClassVar[int]
    ASSISTANT_FIELD_NUMBER: _ClassVar[int]
    DISCONNECTACTION_FIELD_NUMBER: _ClassVar[int]
    HOLDACTION_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTTRANSFERACTION_FIELD_NUMBER: _ClassVar[int]
    PHONECALLTRANSFERACTION_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    error: _common_pb2.Error
    interruption: _common_pb2.AssistantConversationInterruption
    assistant: _common_pb2.AssistantConversationAssistantMessage
    disconnectAction: _common_pb2.AssistantConversationDisconnectAction
    holdAction: _common_pb2.AssistantConverstationHoldAction
    assistantTransferAction: _common_pb2.AssistantConversationAssistantTransferAction
    phoneCallTransferAction: _common_pb2.AssistantConversationPhoneCallTransferAction
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ..., interruption: _Optional[_Union[_common_pb2.AssistantConversationInterruption, _Mapping]] = ..., assistant: _Optional[_Union[_common_pb2.AssistantConversationAssistantMessage, _Mapping]] = ..., disconnectAction: _Optional[_Union[_common_pb2.AssistantConversationDisconnectAction, _Mapping]] = ..., holdAction: _Optional[_Union[_common_pb2.AssistantConverstationHoldAction, _Mapping]] = ..., assistantTransferAction: _Optional[_Union[_common_pb2.AssistantConversationAssistantTransferAction, _Mapping]] = ..., phoneCallTransferAction: _Optional[_Union[_common_pb2.AssistantConversationPhoneCallTransferAction, _Mapping]] = ...) -> None: ...
