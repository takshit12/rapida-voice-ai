from google.protobuf import timestamp_pb2 as _timestamp_pb2
import app.bridges.artifacts.protos.common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class NotificationSetting(_message.Message):
    __slots__ = ("id", "authUserId", "channel", "eventType", "enabled", "status", "createdDate", "updatedDate")
    ID_FIELD_NUMBER: _ClassVar[int]
    AUTHUSERID_FIELD_NUMBER: _ClassVar[int]
    CHANNEL_FIELD_NUMBER: _ClassVar[int]
    EVENTTYPE_FIELD_NUMBER: _ClassVar[int]
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    CREATEDDATE_FIELD_NUMBER: _ClassVar[int]
    UPDATEDDATE_FIELD_NUMBER: _ClassVar[int]
    id: int
    authUserId: int
    channel: str
    eventType: str
    enabled: bool
    status: str
    createdDate: _timestamp_pb2.Timestamp
    updatedDate: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[int] = ..., authUserId: _Optional[int] = ..., channel: _Optional[str] = ..., eventType: _Optional[str] = ..., enabled: bool = ..., status: _Optional[str] = ..., createdDate: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updatedDate: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class UpdateNotificationSettingRequest(_message.Message):
    __slots__ = ("settings",)
    SETTINGS_FIELD_NUMBER: _ClassVar[int]
    settings: _containers.RepeatedCompositeFieldContainer[NotificationSetting]
    def __init__(self, settings: _Optional[_Iterable[_Union[NotificationSetting, _Mapping]]] = ...) -> None: ...

class GetNotificationSettingRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class NotificationSettingResponse(_message.Message):
    __slots__ = ("code", "success", "data", "error")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: _containers.RepeatedCompositeFieldContainer[NotificationSetting]
    error: _common_pb2.Error
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Iterable[_Union[NotificationSetting, _Mapping]]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ...) -> None: ...
