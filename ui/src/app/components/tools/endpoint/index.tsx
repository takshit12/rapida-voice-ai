import { FC, useState, useCallback } from 'react';
import { Endpoint } from '@rapidaai/react';
import { ArrowRight, Plus, Trash2 } from 'lucide-react';
import { cn } from '@/utils';
import { EndpointDropdown } from '@/app/components/dropdown/endpoint-dropdown';
import { FormLabel } from '@/app/components/form-label';
import {
  IBlueBorderButton,
  IRedBorderButton,
} from '@/app/components/form/button';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { Select } from '@/app/components/form/select';
import { InputGroup } from '@/app/components/input-group';
import {
  ConfigureToolProps,
  ToolDefinitionForm,
  TypeKeySelector,
  useParameterManager,
  parseJsonParameters,
  stringifyParameters,
  ParameterType,
  KeyValueParameter,
} from '../common';

// ============================================================================
// Constants
// ============================================================================

const ENDPOINT_PARAMETER_TYPE_OPTIONS: Array<{ name: string; value: string }> =
  [
    { name: 'Tool', value: 'tool' },
    { name: 'Assistant', value: 'assistant' },
    { name: 'Conversation', value: 'conversation' },
    { name: 'Argument', value: 'argument' },
    { name: 'Metadata', value: 'metadata' },
    { name: 'Option', value: 'option' },
  ];

// ============================================================================
// Main Component
// ============================================================================

export const ConfigureEndpoint: FC<ConfigureToolProps> = ({
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

  return (
    <>
      <InputGroup title="Action Definition">
        <div className="flex flex-col gap-8 max-w-6xl">
          <EndpointDropdown
            className={cn('bg-light-background', inputClass)}
            currentEndpoint={getParamValue('tool.endpoint_id')}
            onChangeEndpoint={(endpoint: Endpoint) => {
              if (endpoint) {
                updateParameter('tool.endpoint_id', endpoint.getId());
              }
            }}
          />
          <EndpointArgumentEditor
            endpointParameters={getParamValue('tool.parameters')}
            setEndpointParameters={value =>
              updateParameter('tool.parameters', value)
            }
            inputClass={inputClass}
          />
        </div>
      </InputGroup>

      {toolDefinition && onChangeToolDefinition && (
        <ToolDefinitionForm
          toolDefinition={toolDefinition}
          onChangeToolDefinition={onChangeToolDefinition}
          inputClass={inputClass}
        />
      )}
    </>
  );
};

// ============================================================================
// Endpoint Argument Editor
// ============================================================================

interface EndpointArgumentEditorProps {
  inputClass?: string;
  endpointParameters: string;
  setEndpointParameters: (value: string) => void;
}

const EndpointArgumentEditor: FC<EndpointArgumentEditorProps> = ({
  endpointParameters,
  setEndpointParameters,
  inputClass,
}) => {
  const [params, setParams] = useState<KeyValueParameter[]>(() =>
    parseJsonParameters(endpointParameters),
  );

  const updateParams = useCallback(
    (newParams: KeyValueParameter[]) => {
      setParams(newParams);
      setEndpointParameters(stringifyParameters(newParams));
    },
    [setEndpointParameters],
  );

  const handleTypeChange = useCallback(
    (index: number, newType: string) => {
      const newParams = [...params];
      newParams[index] = { key: `${newType}.`, value: '' };
      updateParams(newParams);
    },
    [params, updateParams],
  );

  const handleKeyChange = useCallback(
    (index: number, newKey: string) => {
      const newParams = [...params];
      const [type] = params[index].key.split('.');
      newParams[index] = { ...params[index], key: `${type}.${newKey}` };
      updateParams(newParams);
    },
    [params, updateParams],
  );

  const handleValueChange = useCallback(
    (index: number, newValue: string) => {
      const newParams = [...params];
      newParams[index] = { ...params[index], value: newValue };
      updateParams(newParams);
    },
    [params, updateParams],
  );

  const handleRemove = useCallback(
    (index: number) => {
      updateParams(params.filter((_, i) => i !== index));
    },
    [params, updateParams],
  );

  const handleAdd = useCallback(() => {
    updateParams([...params, { key: 'assistant.', value: '' }]);
  }, [params, updateParams]);

  return (
    <FieldSet>
      <FormLabel>Parameters ({params.length})</FormLabel>
      <div className="text-sm grid w-full">
        {params.map(({ key, value }, index) => {
          const [type, paramKey] = key.split('.');
          return (
            <EndpointParameterRow
              key={index}
              type={type as ParameterType}
              paramKey={paramKey}
              value={value}
              inputClass={inputClass}
              onTypeChange={newType => handleTypeChange(index, newType)}
              onKeyChange={newKey => handleKeyChange(index, newKey)}
              onValueChange={newValue => handleValueChange(index, newValue)}
              onRemove={() => handleRemove(index)}
            />
          );
        })}
      </div>
      <IBlueBorderButton
        onClick={handleAdd}
        className="justify-between space-x-8"
      >
        <span>Add parameters</span>
        <Plus className="h-4 w-4 ml-1.5" />
      </IBlueBorderButton>
    </FieldSet>
  );
};

// ============================================================================
// Endpoint Parameter Row
// ============================================================================

interface EndpointParameterRowProps {
  type: ParameterType;
  paramKey: string;
  value: string;
  inputClass?: string;
  onTypeChange: (type: string) => void;
  onKeyChange: (key: string) => void;
  onValueChange: (value: string) => void;
  onRemove: () => void;
}

const EndpointParameterRow: FC<EndpointParameterRowProps> = ({
  type,
  paramKey,
  value,
  inputClass,
  onTypeChange,
  onKeyChange,
  onValueChange,
  onRemove,
}) => (
  <div className="grid grid-cols-2 border-b border-gray-300 dark:border-gray-700">
    <div className="flex col-span-1 items-center">
      <Select
        value={type}
        onChange={e => onTypeChange(e.target.value)}
        className={cn('bg-light-background border-none', inputClass)}
        options={ENDPOINT_PARAMETER_TYPE_OPTIONS}
      />
      <TypeKeySelector
        type={type}
        inputClass={inputClass}
        value={paramKey}
        onChange={onKeyChange}
      />
      <div
        className={cn(
          'bg-light-background dark:bg-gray-950 h-full flex items-center justify-center',
          inputClass,
        )}
      >
        <ArrowRight strokeWidth={1.5} className="w-4 h-4" />
      </div>
    </div>
    <div className="col-span-1 flex">
      <Input
        value={value}
        onChange={e => onValueChange(e.target.value)}
        placeholder="Value"
        className={cn('bg-light-background w-full border-none', inputClass)}
      />
      <IRedBorderButton
        className="border-none outline-hidden h-10"
        onClick={onRemove}
        type="button"
      >
        <Trash2 className="w-4 h-4" strokeWidth={1.5} />
      </IRedBorderButton>
    </div>
  </div>
);
