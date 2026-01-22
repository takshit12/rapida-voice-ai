import { FC } from 'react';
import { ConfigureToolProps, ToolDefinitionForm } from '../common';

// ============================================================================
// Main Component
// ============================================================================

export const ConfigureEndOfConversation: FC<ConfigureToolProps> = ({
  inputClass,
  toolDefinition,
  onChangeToolDefinition,
}) => (
  <>
    {toolDefinition && onChangeToolDefinition && (
      <ToolDefinitionForm
        toolDefinition={toolDefinition}
        onChangeToolDefinition={onChangeToolDefinition}
        inputClass={inputClass}
        documentationUrl="https://doc.rapida.ai/assistants/tools/add-end-of-conversation-tool"
        documentationTitle="Know more about End of Conversation that can be supported by rapida"
      />
    )}
  </>
);
