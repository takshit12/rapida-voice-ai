import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { Slider } from '@/app/components/form/slider';
import { Textarea } from '@/app/components/form/textarea';
import { InputGroup } from '@/app/components/input-group';
import { InputHelper } from '@/app/components/input-helper';
import { cn } from '@/utils';
import { FC } from 'react';

export interface ExperienceConfig {
  greeting?: string;
  messageOnError?: string;
  idealTimeout?: string;
  idealMessage?: string;
  maxCallDuration?: string;
  idleTimeoutBackoffTimes?: string;
}

export const ConfigureExperience: FC<{
  experienceConfig: ExperienceConfig;
  setExperienceConfig: (config: ExperienceConfig) => void;
}> = ({ experienceConfig, setExperienceConfig }) => {
  /**
   *
   * @param newGreeting
   */
  const onChangeGreeting = (newGreeting: string) => {
    setExperienceConfig({ ...experienceConfig, greeting: newGreeting });
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
      className="bg-white dark:bg-gray-900 "
    >
      <div className={cn('flex max-w-3xl')}>
        <div className="flex flex-col space-y-6 w-full">
          <FieldSet>
            <FormLabel>Greeting</FormLabel>
            <Textarea
              row={2}
              className="bg-light-background"
              value={experienceConfig.greeting || ''}
              onChange={e => onChangeGreeting(e.target.value)}
              placeholder={
                'Write a custom greeting message. You can use {{variable}} to include dynamic content.'
              }
            />
          </FieldSet>
        </div>
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
                value={experienceConfig.messageOnError || ''}
                onChange={e => onChangeMessageOnError(e.target.value)}
              />
            </FieldSet>
            <FieldSet>
              <FormLabel>Idle Silence Timeout (minute)</FormLabel>
              <div className="flex space-x-2 justify-center items-center">
                <Slider
                  min={2}
                  max={7}
                  step={1}
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
                (3-10 minute).
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
                  min={5}
                  max={15}
                  step={1}
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
                it should be between 5 and 15 minute.
              </InputHelper>
            </FieldSet>
          </div>
        </div>
      </InputGroup>
    </InputGroup>
  );
};
