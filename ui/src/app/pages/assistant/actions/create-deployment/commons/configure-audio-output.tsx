import { Metadata } from '@rapidaai/react';
import { Dropdown } from '@/app/components/dropdown';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { InputGroup } from '@/app/components/input-group';
import { InputHelper } from '@/app/components/input-helper';
import { TextToSpeechProvider } from '@/app/components/providers/text-to-speech';
import { ConditionalInputGroup } from '@/app/components/conditional-input-group';
import { useCallback } from 'react';
import {
  GetDefaultSpeakerConfig,
  GetDefaultTextToSpeechIfInvalid,
} from '@/app/components/providers/text-to-speech/provider';
import {
  CONJUNCTION_BOUNDARIES,
  PRONUNCIATION_DICTIONARIES,
  SENTENCE_BOUNDRIES,
} from '@/providers';

/**
 *
 */
interface ConfigureAudioOutputProviderProps {
  voiceOutputEnable: boolean;
  onChangeVoiceOutputEnable: (b: boolean) => void;
  audioOutputConfig: { provider: string; parameters: Metadata[] };
  setAudioOutputConfig: (config: {
    provider: string;
    parameters: Metadata[];
  }) => void;
}

/**
 *
 * @param param0
 * @returns
 */
export const ConfigureAudioOutputProvider: React.FC<
  ConfigureAudioOutputProviderProps
> = ({
  audioOutputConfig,
  setAudioOutputConfig,
  voiceOutputEnable,
  onChangeVoiceOutputEnable,
}) => {
  // when change the provider reset the parameters for provider
  const onChangeAudioOutputProvider = (providerName: string) => {
    setAudioOutputConfig({
      provider: providerName,
      parameters: GetDefaultTextToSpeechIfInvalid(
        providerName,
        GetDefaultSpeakerConfig([]),
      ),
    });
  };
  const onChangeAudioOutputParameter = (parameters: Metadata[]) => {
    if (audioOutputConfig)
      setAudioOutputConfig({ ...audioOutputConfig, parameters });
  };

  /**
   * to get parameters
   */
  const getParamValue = useCallback(
    (key: string, defaultValue: any) => {
      const param = audioOutputConfig.parameters?.find(p => p.getKey() === key);
      return param ? param.getValue() : defaultValue;
    },
    [JSON.stringify(audioOutputConfig.parameters)],
  );

  const updateParameter = (key: string, value: string) => {
    const updatedParams = (audioOutputConfig.parameters || []).map(param => {
      if (param.getKey() === key) {
        const updatedParam = new Metadata();
        updatedParam.setKey(key);
        updatedParam.setValue(value);
        return updatedParam;
      }
      return param;
    });
    if (!updatedParams.some(param => param.getKey() === key)) {
      const newParam = new Metadata();
      newParam.setKey(key);
      newParam.setValue(value);
      updatedParams.push(newParam);
    }
    onChangeAudioOutputParameter(updatedParams);
  };

  return (
    <ConditionalInputGroup
      title="Voice Output"
      enable={voiceOutputEnable}
      className="bg-white dark:bg-gray-900"
      onChangeEnable={onChangeVoiceOutputEnable}
    >
      <TextToSpeechProvider
        onChangeProvider={onChangeAudioOutputProvider}
        onChangeParameter={onChangeAudioOutputParameter}
        provider={audioOutputConfig.provider}
        parameters={audioOutputConfig.parameters}
      />
      {audioOutputConfig.provider && (
        <>
          <InputGroup
            initiallyExpanded={false}
            title="Speaker Expereience"
            className="mx-0 my-0 mt-6"
          >
            <div className="space-y-6 w-full max-w-6xl">
              <FieldSet className="relative">
                <FormLabel>Sentence Boundaries</FormLabel>
                <Dropdown
                  multiple
                  className="bg-light-background dark:bg-gray-950 max-w-6xl"
                  currentValue={getParamValue(
                    'speaker.sentence.boundaries',
                    SENTENCE_BOUNDRIES.join('<|||>'),
                  ).split('<|||>')}
                  setValue={v => {
                    updateParameter(
                      'speaker.sentence.boundaries',
                      v.join('<|||>'),
                    );
                  }}
                  allValue={SENTENCE_BOUNDRIES}
                  placeholder="Select all that applies"
                  option={c => {
                    return (
                      <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
                        <span className="truncate capitalize">{c}</span>
                      </span>
                    );
                  }}
                  label={c => {
                    return (
                      <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
                        {c.map(x => {
                          return (
                            <span key={x} className="truncate">
                              {x}
                            </span>
                          );
                        })}
                      </span>
                    );
                  }}
                />
                <InputHelper>
                  These are the sentence that are considered valid boundaries or
                  delimiters. This helps decides the chunks that are sent to the
                  voice provider for the voice generation as the LLM tokens are
                  streaming in
                </InputHelper>
              </FieldSet>

              <FieldSet className="relative">
                <FormLabel>Pronunciation Dictionaries</FormLabel>
                <Dropdown
                  multiple
                  className="bg-light-background dark:bg-gray-950 max-w-6xl"
                  currentValue={getParamValue(
                    'speaker.pronunciation.dictionaries',
                    '',
                  ).split('<|||>')}
                  setValue={v => {
                    updateParameter(
                      'speaker.pronunciation.dictionaries',
                      v.join('<|||>'),
                    );
                  }}
                  allValue={PRONUNCIATION_DICTIONARIES}
                  placeholder="Select all that applies"
                  option={c => {
                    return (
                      <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
                        <span className="truncate capitalize">{c}</span>
                      </span>
                    );
                  }}
                  label={c => {
                    return (
                      <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
                        {c.map(x => {
                          return (
                            <span key={x} className="truncate">
                              {x}
                            </span>
                          );
                        })}
                      </span>
                    );
                  }}
                />
                <InputHelper>
                  Pronunciation dictionaries help define custom pronunciations
                  for words, abbreviations, and acronyms. They ensure correct
                  pronunciation of domain-specific terms, names, or technical
                  jargon that may not be pronounced correctly by default
                  text-to-speech systems.
                </InputHelper>
              </FieldSet>
              <FieldSet className="relative">
                <FormLabel>Conjunction Boundaries</FormLabel>
                <Dropdown
                  multiple
                  className="bg-light-background dark:bg-gray-950 max-w-6xl"
                  currentValue={getParamValue(
                    'speaker.conjunction.boundaries',
                    '',
                  ).split('<|||>')}
                  setValue={v => {
                    updateParameter(
                      'speaker.conjunction.boundaries',
                      v.join('<|||>'),
                    );
                  }}
                  allValue={CONJUNCTION_BOUNDARIES}
                  placeholder="Select all that applies"
                  option={c => {
                    return (
                      <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
                        <span className="truncate capitalize">{c}</span>
                      </span>
                    );
                  }}
                  label={c => {
                    return (
                      <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
                        {c.map(x => {
                          return (
                            <span key={x} className="truncate">
                              {x}
                            </span>
                          );
                        })}
                      </span>
                    );
                  }}
                />
                <InputHelper>
                  These are the punctuations that are considered valid
                  boundaries or delimiters. This helps decides to add pause
                  before delivering to voice provider
                </InputHelper>
              </FieldSet>
              <FieldSet className="col-span-1">
                <FormLabel>Pause duration (Millisecond)</FormLabel>
                <Input
                  min={100}
                  max={300}
                  className="bg-light-background w-16"
                  value={getParamValue('speaker.conjunction.break', '240')}
                  onChange={e =>
                    updateParameter('speaker.conjunction.break', e.target.value)
                  }
                />
              </FieldSet>
            </div>
          </InputGroup>
        </>
      )}
    </ConditionalInputGroup>
  );
};
