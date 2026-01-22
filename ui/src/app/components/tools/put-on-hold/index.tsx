import { FC } from 'react';
import { InfoIcon } from 'lucide-react';
import { cn } from '@/utils';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { Slider } from '@/app/components/form/slider';
import { Tooltip } from '@/app/components/tooltip';
import { InputGroup } from '@/app/components/input-group';
import {
  ConfigureToolProps,
  ToolDefinitionForm,
  useParameterManager,
} from '../common';

// ============================================================================
// Main Component
// ============================================================================

export const ConfigurePutOnHold: FC<ConfigureToolProps> = ({
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
          <div className="grid grid-cols-2 w-full gap-4">
            <FieldSet className="flex justify-between">
              <FormLabel htmlFor="max_hold_time">
                Max hold time second
                <Tooltip icon={<InfoIcon className="w-4 h-4 ml-1" />}>
                  <p className={cn('font-normal text-sm p-1 w-64')}>
                    Maximum hold duration before auto-resume or callback.
                  </p>
                </Tooltip>
              </FormLabel>
              <div className="flex justify-between items-center space-x-2">
                <Slider
                  min={3}
                  max={10}
                  step={1}
                  value={getParamValue('tool.max_hold_time')}
                  onSlide={(value: number) =>
                    updateParameter('tool.max_hold_time', value.toString())
                  }
                />
                <Input
                  id="max_hold_time"
                  className={cn(
                    'py-0 px-1 tabular-nums border w-10 h-6 text-xs',
                    inputClass,
                  )}
                  min={0}
                  max={10}
                  type="number"
                  value={Number(getParamValue('tool.max_hold_time'))}
                  onChange={e =>
                    updateParameter('tool.max_hold_time', e.target.value)
                  }
                />
              </div>
            </FieldSet>
          </div>
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
