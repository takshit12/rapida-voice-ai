import { Dropdown } from '@/app/components/dropdown';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import { InputGroup } from '@/app/components/input-group';
import { HTMLAttributes } from 'react';
import { EndOfSpeech } from '@/providers';
import { cn } from '@/utils';
import { Slider } from '@/app/components/form/slider';
import { Input } from '@/app/components/form/input';
import { InputHelper } from '@/app/components/input-helper';

interface EndOfSpeechProviderProps extends HTMLAttributes<HTMLDivElement> {
  endOfSpeechProvider: string;
  onChangeEndOfSpeechProvider: (string) => void;
  endOfSepeechTimeout: string;
  onChangeEndOfSepeechTimeout: (n: string) => void;
}

export const EndOfSpeechProvider: React.FC<EndOfSpeechProviderProps> = ({
  endOfSpeechProvider,
  onChangeEndOfSpeechProvider,
  endOfSepeechTimeout,
  onChangeEndOfSepeechTimeout,
  className,
}) => {
  return (
    <InputGroup
      initiallyExpanded={false}
      title="End of speech"
      className={cn('bg-white dark:bg-gray-900', className)}
    >
      <div className="space-y-6">
        <FieldSet>
          <FormLabel>End-of-speech provider</FormLabel>
          <Dropdown
            className="bg-light-background max-w-full dark:bg-gray-950"
            currentValue={EndOfSpeech().find(
              x => x.code === endOfSpeechProvider,
            )}
            setValue={v => {
              onChangeEndOfSpeechProvider(v.code);
            }}
            allValue={EndOfSpeech()}
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
          <FormLabel>Activity Timeout</FormLabel>
          <div className="flex space-x-2 justify-center items-center">
            <Slider
              min={500}
              max={4000}
              step={100}
              value={parseInt(endOfSepeechTimeout)}
              onSlide={v => {
                onChangeEndOfSepeechTimeout(v.toString());
              }}
            />
            <Input
              min={500}
              max={4000}
              className="bg-light-background w-16"
              value={endOfSepeechTimeout}
              onChange={e => onChangeEndOfSepeechTimeout(e.target.value)}
            />
          </div>
          <InputHelper>
            Duration of silence after which Rapida starts finalizing a message
            EOS: Based on silence and max time (500-4000ms).
          </InputHelper>
        </FieldSet>
      </div>
    </InputGroup>
  );
};
