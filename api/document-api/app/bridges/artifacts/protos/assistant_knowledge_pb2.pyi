from google.protobuf import timestamp_pb2 as _timestamp_pb2
import app.bridges.artifacts.protos.common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AssistantKnowledge(_message.Message):
    __slots__ = ("id", "knowledgeId", "rerankerEnable", "topK", "scoreThreshold", "knowledge", "retrievalMethod", "rerankerModelProviderId", "rerankerModelProviderName", "assistantKnowledgeRerankerOptions", "createdDate", "updatedDate", "status")
    ID_FIELD_NUMBER: _ClassVar[int]
    KNOWLEDGEID_FIELD_NUMBER: _ClassVar[int]
    RERANKERENABLE_FIELD_NUMBER: _ClassVar[int]
    TOPK_FIELD_NUMBER: _ClassVar[int]
    SCORETHRESHOLD_FIELD_NUMBER: _ClassVar[int]
    KNOWLEDGE_FIELD_NUMBER: _ClassVar[int]
    RETRIEVALMETHOD_FIELD_NUMBER: _ClassVar[int]
    RERANKERMODELPROVIDERID_FIELD_NUMBER: _ClassVar[int]
    RERANKERMODELPROVIDERNAME_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTKNOWLEDGERERANKEROPTIONS_FIELD_NUMBER: _ClassVar[int]
    CREATEDDATE_FIELD_NUMBER: _ClassVar[int]
    UPDATEDDATE_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    id: int
    knowledgeId: int
    rerankerEnable: bool
    topK: int
    scoreThreshold: float
    knowledge: _common_pb2.Knowledge
    retrievalMethod: str
    rerankerModelProviderId: int
    rerankerModelProviderName: str
    assistantKnowledgeRerankerOptions: _containers.RepeatedCompositeFieldContainer[_common_pb2.Metadata]
    createdDate: _timestamp_pb2.Timestamp
    updatedDate: _timestamp_pb2.Timestamp
    status: str
    def __init__(self, id: _Optional[int] = ..., knowledgeId: _Optional[int] = ..., rerankerEnable: bool = ..., topK: _Optional[int] = ..., scoreThreshold: _Optional[float] = ..., knowledge: _Optional[_Union[_common_pb2.Knowledge, _Mapping]] = ..., retrievalMethod: _Optional[str] = ..., rerankerModelProviderId: _Optional[int] = ..., rerankerModelProviderName: _Optional[str] = ..., assistantKnowledgeRerankerOptions: _Optional[_Iterable[_Union[_common_pb2.Metadata, _Mapping]]] = ..., createdDate: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updatedDate: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., status: _Optional[str] = ...) -> None: ...

class CreateAssistantKnowledgeRequest(_message.Message):
    __slots__ = ("knowledgeId", "assistantId", "rerankerModelProviderId", "rerankerModelProviderName", "assistantKnowledgeRerankerOptions", "topK", "scoreThreshold", "retrievalMethod", "rerankerEnable")
    KNOWLEDGEID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    RERANKERMODELPROVIDERID_FIELD_NUMBER: _ClassVar[int]
    RERANKERMODELPROVIDERNAME_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTKNOWLEDGERERANKEROPTIONS_FIELD_NUMBER: _ClassVar[int]
    TOPK_FIELD_NUMBER: _ClassVar[int]
    SCORETHRESHOLD_FIELD_NUMBER: _ClassVar[int]
    RETRIEVALMETHOD_FIELD_NUMBER: _ClassVar[int]
    RERANKERENABLE_FIELD_NUMBER: _ClassVar[int]
    knowledgeId: int
    assistantId: int
    rerankerModelProviderId: int
    rerankerModelProviderName: str
    assistantKnowledgeRerankerOptions: _containers.RepeatedCompositeFieldContainer[_common_pb2.Metadata]
    topK: int
    scoreThreshold: float
    retrievalMethod: str
    rerankerEnable: bool
    def __init__(self, knowledgeId: _Optional[int] = ..., assistantId: _Optional[int] = ..., rerankerModelProviderId: _Optional[int] = ..., rerankerModelProviderName: _Optional[str] = ..., assistantKnowledgeRerankerOptions: _Optional[_Iterable[_Union[_common_pb2.Metadata, _Mapping]]] = ..., topK: _Optional[int] = ..., scoreThreshold: _Optional[float] = ..., retrievalMethod: _Optional[str] = ..., rerankerEnable: bool = ...) -> None: ...

class UpdateAssistantKnowledgeRequest(_message.Message):
    __slots__ = ("id", "knowledgeId", "assistantId", "rerankerModelProviderId", "rerankerModelProviderName", "assistantKnowledgeRerankerOptions", "scoreThreshold", "topK", "retrievalMethod", "rerankerEnable")
    ID_FIELD_NUMBER: _ClassVar[int]
    KNOWLEDGEID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    RERANKERMODELPROVIDERID_FIELD_NUMBER: _ClassVar[int]
    RERANKERMODELPROVIDERNAME_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTKNOWLEDGERERANKEROPTIONS_FIELD_NUMBER: _ClassVar[int]
    SCORETHRESHOLD_FIELD_NUMBER: _ClassVar[int]
    TOPK_FIELD_NUMBER: _ClassVar[int]
    RETRIEVALMETHOD_FIELD_NUMBER: _ClassVar[int]
    RERANKERENABLE_FIELD_NUMBER: _ClassVar[int]
    id: int
    knowledgeId: int
    assistantId: int
    rerankerModelProviderId: int
    rerankerModelProviderName: str
    assistantKnowledgeRerankerOptions: _containers.RepeatedCompositeFieldContainer[_common_pb2.Metadata]
    scoreThreshold: float
    topK: int
    retrievalMethod: str
    rerankerEnable: bool
    def __init__(self, id: _Optional[int] = ..., knowledgeId: _Optional[int] = ..., assistantId: _Optional[int] = ..., rerankerModelProviderId: _Optional[int] = ..., rerankerModelProviderName: _Optional[str] = ..., assistantKnowledgeRerankerOptions: _Optional[_Iterable[_Union[_common_pb2.Metadata, _Mapping]]] = ..., scoreThreshold: _Optional[float] = ..., topK: _Optional[int] = ..., retrievalMethod: _Optional[str] = ..., rerankerEnable: bool = ...) -> None: ...

class GetAssistantKnowledgeRequest(_message.Message):
    __slots__ = ("id", "assistantId")
    ID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    id: int
    assistantId: int
    def __init__(self, id: _Optional[int] = ..., assistantId: _Optional[int] = ...) -> None: ...

class DeleteAssistantKnowledgeRequest(_message.Message):
    __slots__ = ("id", "assistantId")
    ID_FIELD_NUMBER: _ClassVar[int]
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    id: int
    assistantId: int
    def __init__(self, id: _Optional[int] = ..., assistantId: _Optional[int] = ...) -> None: ...

class GetAssistantKnowledgeResponse(_message.Message):
    __slots__ = ("code", "success", "data", "error")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: AssistantKnowledge
    error: _common_pb2.Error
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Union[AssistantKnowledge, _Mapping]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ...) -> None: ...

class GetAllAssistantKnowledgeRequest(_message.Message):
    __slots__ = ("assistantId", "paginate", "criterias")
    ASSISTANTID_FIELD_NUMBER: _ClassVar[int]
    PAGINATE_FIELD_NUMBER: _ClassVar[int]
    CRITERIAS_FIELD_NUMBER: _ClassVar[int]
    assistantId: int
    paginate: _common_pb2.Paginate
    criterias: _containers.RepeatedCompositeFieldContainer[_common_pb2.Criteria]
    def __init__(self, assistantId: _Optional[int] = ..., paginate: _Optional[_Union[_common_pb2.Paginate, _Mapping]] = ..., criterias: _Optional[_Iterable[_Union[_common_pb2.Criteria, _Mapping]]] = ...) -> None: ...

class GetAllAssistantKnowledgeResponse(_message.Message):
    __slots__ = ("code", "success", "data", "error", "paginated")
    CODE_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    PAGINATED_FIELD_NUMBER: _ClassVar[int]
    code: int
    success: bool
    data: _containers.RepeatedCompositeFieldContainer[AssistantKnowledge]
    error: _common_pb2.Error
    paginated: _common_pb2.Paginated
    def __init__(self, code: _Optional[int] = ..., success: bool = ..., data: _Optional[_Iterable[_Union[AssistantKnowledge, _Mapping]]] = ..., error: _Optional[_Union[_common_pb2.Error, _Mapping]] = ..., paginated: _Optional[_Union[_common_pb2.Paginated, _Mapping]] = ...) -> None: ...
