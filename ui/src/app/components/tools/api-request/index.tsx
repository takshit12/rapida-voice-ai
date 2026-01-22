import { FC, useState, useCallback } from 'react';
import { ArrowRight, Plus, Trash2 } from 'lucide-react';
import { cn } from '@/utils';
import { FormLabel } from '@/app/components/form-label';
import {
  IBlueBorderButton,
  IRedBorderButton,
} from '@/app/components/form/button';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { Select } from '@/app/components/form/select';
import { InputGroup } from '@/app/components/input-group';
import { APiStringHeader } from '@/app/components/external-api/api-header';
import {
  ConfigureToolProps,
  ToolDefinitionForm,
  TypeKeySelector,
  useParameterManager,
  parseJsonParameters,
  stringifyParameters,
  PARAMETER_TYPE_OPTIONS,
  HTTP_METHOD_OPTIONS,
  ParameterType,
  KeyValueParameter,
} from '../common';

// ============================================================================
// Main Component
// ============================================================================

export const ConfigureAPIRequest: FC<ConfigureToolProps> = ({
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
          <div className="flex space-x-2">
            <FieldSet className="relative w-40">
              <FormLabel>Method</FormLabel>
              <Select
                value={getParamValue('tool.method')}
                onChange={e => updateParameter('tool.method', e.target.value)}
                className={cn('bg-light-background', inputClass)}
                options={HTTP_METHOD_OPTIONS}
              />
            </FieldSet>
            <FieldSet className="relative w-full">
              <FormLabel>Server Url</FormLabel>
              <Input
                value={getParamValue('tool.endpoint')}
                onChange={e => updateParameter('tool.endpoint', e.target.value)}
                placeholder="https://your-domain.com/api/v1/resource"
                className={cn('bg-light-background', inputClass)}
              />
            </FieldSet>
          </div>

          <FieldSet>
            <FormLabel>Headers</FormLabel>
            <APiStringHeader
              inputClass={inputClass}
              headerValue={getParamValue('tool.headers')}
              setHeaderValue={value => updateParameter('tool.headers', value)}
            />
          </FieldSet>

          <ApiParameterEditor
            inputClass={inputClass}
            apiParameters={getParamValue('tool.parameters')}
            setApiParameters={value =>
              updateParameter('tool.parameters', value)
            }
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
// API Parameter Editor
// ============================================================================

interface ApiParameterEditorProps {
  inputClass?: string;
  apiParameters: string;
  setApiParameters: (value: string) => void;
}

const ApiParameterEditor: FC<ApiParameterEditorProps> = ({
  apiParameters,
  setApiParameters,
  inputClass,
}) => {
  const [params, setParams] = useState<KeyValueParameter[]>(() =>
    parseJsonParameters(apiParameters),
  );

  const updateParams = useCallback(
    (newParams: KeyValueParameter[]) => {
      setParams(newParams);
      setApiParameters(stringifyParameters(newParams));
    },
    [setApiParameters],
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
            <ParameterRow
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
// Parameter Row
// ============================================================================

interface ParameterRowProps {
  type: ParameterType;
  paramKey: string;
  value: string;
  inputClass?: string;
  onTypeChange: (type: string) => void;
  onKeyChange: (key: string) => void;
  onValueChange: (value: string) => void;
  onRemove: () => void;
}

const ParameterRow: FC<ParameterRowProps> = ({
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
        options={[...PARAMETER_TYPE_OPTIONS]}
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
