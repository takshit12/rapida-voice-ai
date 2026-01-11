import { Dropdown } from '@/app/components/dropdown';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import { InputGroup } from '@/app/components/input-group';
import { HTMLAttributes } from 'react';
import { VAD } from '@/providers';
import { cn } from '@/utils';
import { Slider } from '@/app/components/form/slider';
import { Input } from '@/app/components/form/input';
import { InputHelper } from '@/app/components/input-helper';

interface VADProviderProps extends HTMLAttributes<HTMLDivElement> {
  vadProvider: string;
  onChangeVADProvider: (string) => void;
  vadThreshold: string;
  onChangeVadThreshold: (n: string) => void;
}

export const VADProvider: React.FC<VADProviderProps> = ({
  vadProvider,
  onChangeVADProvider,
  vadThreshold,
  onChangeVadThreshold,
  className,
}) => {
  return (
    <InputGroup
      title="VAD"
      initiallyExpanded={false}
      className={cn('bg-white dark:bg-gray-900', className)}
    >
      <div className="space-y-6">
        <FieldSet>
          <FormLabel>VAD provider</FormLabel>
          <Dropdown
            className="bg-light-background max-w-full dark:bg-gray-950"
            currentValue={VAD().find(x => x.code === vadProvider)}
            setValue={v => {
              onChangeVADProvider(v.code);
            }}
            allValue={VAD()}
            placeholder="Select end of speech provider"
            option={c => {
              return (
                <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
                  <span className="truncate capitalize">{c.name}</span>
                </span>
              );
            }}
            label={c => {
              return (
                <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
                  <span className="truncate capitalize">{c.name}</span>
                </span>
              );
            }}
          />
        </FieldSet>
        <FieldSet className="col-span-1">
          <FormLabel>VAD Threshold</FormLabel>
          <div className="flex space-x-2 justify-center items-center">
            <Slider
              min={0.5}
              max={1}
              step={0.01}
              value={parseFloat(vadThreshold)}
              onSlide={v => {
                onChangeVadThreshold(v.toString());
              }}
            />
            <Input
              min={0.1}
              max={1}
              className="bg-light-background w-16"
              value={vadThreshold}
              onChange={e => onChangeVadThreshold(e.target.value)}
            />
          </div>
          <InputHelper>
            The probability threshold above which we detect speech. A good
            default is 0.8.
          </InputHelper>
        </FieldSet>
      </div>
    </InputGroup>
  );
};
