"""
Copyright (c) 2023-2025 RapidaAI
Author: Prashant Srivastav <prashant@rapida.ai>

Licensed under GPL-2.0 with Rapida Additional Terms.
See LICENSE.md for details or contact sales@rapida.ai for commercial use.
"""
import logging

from google.protobuf.json_format import ParseDict
from grpc.aio import Metadata

from app.bridges import GRPCBridge
from app.bridges.artifacts.protos import (
    vault_api_pb2,
    vault_api_pb2_grpc,
)
from app.exceptions.bridges_exceptions import BridgeException

_log = logging.getLogger("bridges.vault_bridge")


class VaultBridge(GRPCBridge):
    
    async def get_credential(
        self, auth_token: str, crendential_id: int
    ) -> "vault_api_pb2.VaultCredential":
        """
        The function `get_credential` retrieves provider credentials using gRPC communication
        with metadata authentication.
        :param auth_token: The `auth_token` parameter is a string that represents the authentication token
        used to authenticate the request to retrieve provider credentials from the vault. It is typically
        a secure token that grants access to the vault service
        :type auth_token: str
        :return: The `get_provider_credential` method returns an object of type
        `vault_api_pb2.VaultCredential`, which represents the credential information retrieved from the
        Vault service for a specific provider and organization.
        """
        # metadata for request
        _metadata: Metadata = Metadata()
        _metadata.add("x-internal-service-key", auth_token)

        request = vault_api_pb2.GetCredentialRequest()
        request.vaultId = crendential_id
        response = await self.fetch(
            stub=vault_api_pb2_grpc.VaultServiceStub,
            attr="GetCredential",
            message_type=request,
            preserving_proto_field_name=True,
            metadata=_metadata,
        )

        result = ParseDict(response, vault_api_pb2.GetCredentialResponse())
        if not result or not result.success:
            raise BridgeException(
                message="Unable to receive provider credentials.", bridge_name="vault"
            )
        return result.data
