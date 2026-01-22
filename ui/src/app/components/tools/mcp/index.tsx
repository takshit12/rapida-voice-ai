import { FC } from 'react';
import { cn } from '@/utils';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { Textarea } from '@/app/components/form/textarea';
import { InputGroup } from '@/app/components/input-group';
import {
  ConfigureToolProps,
  useParameterManager,
} from '../common';
import { BlueNoticeBlock } from '@/app/components/container/message/notice-block';

// ============================================================================
// Main Component
// ============================================================================

export const ConfigureMCP: FC<ConfigureToolProps> = ({
  toolDefinition,
  onChangeToolDefinition,
  onParameterChange,
  parameters,
  inputClass,
}) => {
  const { getParamValue, updateParameter } = useParameterManager(
    parameters,
    onParameterChange,
  );

  const serverUrl = getParamValue('mcp.server_url');

  const handleChange = (
    field: 'name' | 'description',
    value: string,
  ) => {
    if (toolDefinition && onChangeToolDefinition) {
      onChangeToolDefinition({ ...toolDefinition, [field]: value });
    }
  };

  return (
    <>
      <InputGroup title="MCP Tool Configuration">
        <div className="flex flex-col gap-6 max-w-6xl">
          <FieldSet>
            <FormLabel>Name</FormLabel>
            <Input
              className={cn('bg-light-background', inputClass)}
              value={toolDefinition?.name || ''}
              onChange={e => handleChange('name', e.target.value)}
              placeholder="Enter MCP tool name"
            />
          </FieldSet>

          <FieldSet>
            <FormLabel>Description</FormLabel>
            <Textarea
              className={cn('bg-light-background', inputClass)}
              value={toolDefinition?.description || ''}
              onChange={e => handleChange('description', e.target.value)}
              placeholder="A tool description or definition of when this MCP tool will get triggered."
              rows={2}
            />
          </FieldSet>

          <FieldSet>
            <FormLabel>MCP Server URL</FormLabel>
            <Input
              className={cn('bg-light-background', inputClass)}
              value={serverUrl}
              onChange={e => updateParameter('mcp.server_url', e.target.value)}
              placeholder="https://your-mcp-server.com"
              type="url"
            />
          </FieldSet>

          <BlueNoticeBlock>
            <div className="text-sm text-blue-900 dark:text-blue-100">
              <div className="text-blue-700 dark:text-blue-300">
                The tool definitions will be automatically fetched from the MCP
                server when the assistant initializes via Server-Sent Events
                (SSE). The name and description you provide above will be used
                to identify this MCP server.
              </div>
            </div>
          </BlueNoticeBlock>
        </div>
      </InputGroup>
    </>
  );
};
