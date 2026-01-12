import ConfigSelect from '@/app/components/configuration/config-var/config-select';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { Slider } from '@/app/components/form/slider';
import { Textarea } from '@/app/components/form/textarea';
import { InputGroup } from '@/app/components/input-group';
import { InputHelper } from '@/app/components/input-helper';
import { ExperienceConfig } from '@/app/pages/assistant/actions/create-deployment/commons/configure-experience';
import { cn } from '@/utils';
import { FC } from 'react';

export interface WebWidgetExperienceConfig extends ExperienceConfig {
  suggestions: string[];
}

export const ConfigureExperience: FC<{
  experienceConfig: WebWidgetExperienceConfig;
  setExperienceConfig: (config: WebWidgetExperienceConfig) => void;
}> = ({ experienceConfig, setExperienceConfig }) => {
  const { greeting, suggestions, messageOnError } = experienceConfig;

  /**
   *
   * @param newGreeting
   */
  const onChangeGreeting = (newGreeting: string) => {
    setExperienceConfig({ ...experienceConfig, greeting: newGreeting });
  };

  const onChangeSuggestions = (newSuggestions: string[]) => {
    setExperienceConfig({ ...experienceConfig, suggestions: newSuggestions });
  };

  const onChangeMessageOnError = (newMessageOnError: string) => {
    setExperienceConfig({
      ...experienceConfig,
      messageOnError: newMessageOnError,
    });
  };

  const onChangeIdealMessage = (idealMessage: string) => {
    setExperienceConfig({
      ...experienceConfig,
      idealMessage: idealMessage,
    });
  };

  const onChangeIdealTimeout = (idealTimeout: string) => {
    setExperienceConfig({
      ...experienceConfig,
      idealTimeout: idealTimeout,
    });
  };
  const onChangeMaxCallDuration = (duration: string) => {
    setExperienceConfig({
      ...experienceConfig,
      maxCallDuration: duration,
    });
  };

  const onChangeIdleTimeoutBackoffTimes = (no: string) => {
    setExperienceConfig({
      ...experienceConfig,
      idleTimeoutBackoffTimes: no,
    });
  };

  return (
    <InputGroup
      title="General Experience"
      className="bg-white dark:bg-gray-900"
    >
      <div className="flex flex-col space-y-6">
        <FieldSet className="block flex-1 md:col-span-2">
          <FormLabel>Greeting</FormLabel>
          <Textarea
            row={2}
            value={greeting}
            className="bg-light-background"
            onChange={e => onChangeGreeting(e.target.value)}
            placeholder={
              'Describe your agent so that users know how to use it. This will appear as a welcome message.'
            }
          />
        </FieldSet>
        <FieldSet className="block flex-1 md:col-span-2">
          <FormLabel>Quickstart questions</FormLabel>
          <ConfigSelect
            options={suggestions}
            label="Add new questions"
            placeholder="Add frequently asked question."
            onChange={onChangeSuggestions}
          />
        </FieldSet>
      </div>
      <InputGroup
        title={'Advanced Experience Configuration'}
        className="mx-0 my-0 mt-6"
        initiallyExpanded={false}
      >
        <div className={cn('flex max-w-3xl')}>
          <div className="flex flex-col space-y-6 w-full">
            <FieldSet>
              <FormLabel>Error Message</FormLabel>
              <Input
                className="bg-light-background"
                placeholder="Message that will be send to the user when error occured"
                value={messageOnError}
                onChange={e => onChangeMessageOnError(e.target.value)}
              />
            </FieldSet>
            <FieldSet>
              <FormLabel>Idle Silence Timeout (millisecond)</FormLabel>
              <div className="flex space-x-2 justify-center items-center">
                <Slider
                  min={3000}
                  max={10000}
                  step={500}
                  value={
                    experienceConfig.idealTimeout &&
                    parseInt(experienceConfig.idealTimeout)
                  }
                  onSlide={(v: number) => {
                    onChangeIdealTimeout(v.toString());
                  }}
                />
                <Input
                  className="bg-light-background w-16"
                  value={experienceConfig.idealTimeout}
                  onChange={e => {
                    onChangeIdealTimeout(e.target.value);
                  }}
                />
              </div>
              <InputHelper>
                Duration of silence after which Rapida will interrupt the user
                (3000-10000ms).
              </InputHelper>
            </FieldSet>
            <FieldSet>
              <FormLabel>Idle Timeout Backoff (Times)</FormLabel>
              <div className="flex space-x-2 justify-center items-center">
                <Slider
                  min={0}
                  max={5}
                  step={1}
                  value={
                    experienceConfig.idleTimeoutBackoffTimes &&
                    parseInt(experienceConfig.idleTimeoutBackoffTimes)
                  }
                  onSlide={(v: number) => {
                    onChangeIdleTimeoutBackoffTimes(v.toString());
                  }}
                />
                <Input
                  className="bg-light-background w-16"
                  value={experienceConfig.idleTimeoutBackoffTimes}
                  onChange={e => {
                    onChangeIdleTimeoutBackoffTimes(e.target.value);
                  }}
                />
              </div>
              <InputHelper>
                Number of times the idle timeout duration increases after it
                triggers. Each time adds the base timeout again (e.g. 3 → 6 → 9
                minutes).
              </InputHelper>
            </FieldSet>
            <FieldSet className="relative col-span-1">
              <FormLabel>Idle Message</FormLabel>
              <Input
                className="bg-light-background"
                placeholder="Message that the assistant will speak when the user hasn't responded."
                value={experienceConfig.idealMessage}
                onChange={e => onChangeIdealMessage(e.target.value)}
              />
              <InputHelper>
                Message that the assistant will speak when the user hasn't
                responded.
              </InputHelper>
            </FieldSet>
            <FieldSet>
              <FormLabel>Maximum Session Duration (millisecond)</FormLabel>
              <div className="flex space-x-2 justify-center items-center">
                <Slider
                  min={5000}
                  max={15000}
                  step={1000}
                  value={
                    experienceConfig.maxCallDuration &&
                    parseInt(experienceConfig.maxCallDuration)
                  }
                  onSlide={(v: number) => {
                    onChangeMaxCallDuration(v.toString());
                  }}
                />
                <Input
                  className="bg-light-background w-16"
                  value={experienceConfig.maxCallDuration}
                  onChange={e => {
                    onChangeMaxCallDuration(e.target.value);
                  }}
                />
              </div>
              <InputHelper>
                Maximum Session Duration. Set the time limit for sessions values
                it should be between 5000ms and 15000ms.
              </InputHelper>
            </FieldSet>
          </div>
        </div>
      </InputGroup>
    </InputGroup>
  );
};
