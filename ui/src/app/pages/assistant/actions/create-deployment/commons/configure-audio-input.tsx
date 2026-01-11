import { useCallback } from 'react';
import { SpeechToTextProvider } from '@/app/components/providers/speech-to-text';
import { ConditionalInputGroup } from '@/app/components/conditional-input-group';
import { NoiseCancellationProvider } from '@/app/components/providers/noise-removal';
import { EndOfSpeechProvider } from '@/app/components/providers/end-of-speech';
import { Metadata } from '@rapidaai/react';
import {
  GetDefaultMicrophoneConfig,
  GetDefaultSpeechToTextIfInvalid,
} from '@/app/components/providers/speech-to-text/provider';
import { VADProvider } from '@/app/components/providers/vad';

/**
 *
 */
interface ConfigureAudioInputProviderProps {
  voiceInputEnable: boolean;
  onChangeVoiceInputEnable: (b: boolean) => void;
  audioInputConfig: { provider: string; parameters: Metadata[] };
  setAudioInputConfig: (config: {
    provider: string;
    parameters: Metadata[];
  }) => void;
}

/**
 *
 * @param param0
 * @returns
 */
export const ConfigureAudioInputProvider: React.FC<
  ConfigureAudioInputProviderProps
> = ({
  voiceInputEnable,
  onChangeVoiceInputEnable,
  audioInputConfig,
  setAudioInputConfig,
}) => {
  // when changing the provide then reset the config
  const onChangeAudioInputProvider = (providerName: string) => {
    setAudioInputConfig({
      provider: providerName,
      parameters: GetDefaultSpeechToTextIfInvalid(
        providerName,
        GetDefaultMicrophoneConfig([]),
      ),
    });
  };

  //   change the parameter for audio input config
  const onChangeAudioInputParameter = (parameters: Metadata[]) => {
    if (audioInputConfig)
      setAudioInputConfig({ ...audioInputConfig, parameters });
  };

  /**
   * to get parameters
   */
  const getParamValue = useCallback(
    (key: string, defaultValue: any) => {
      const param = audioInputConfig.parameters?.find(p => p.getKey() === key);
      return param ? param.getValue() : defaultValue;
    },
    [JSON.stringify(audioInputConfig.parameters)],
  );

  //   update the parameters
  const updateParameter = (key: string, value: string) => {
    const updatedParams = (audioInputConfig.parameters || []).map(param => {
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
    onChangeAudioInputParameter(updatedParams);
  };

  return (
    <ConditionalInputGroup
      title="Voice Input"
      className="my-0 bg-white dark:bg-gray-900"
      enable={voiceInputEnable}
      onChangeEnable={onChangeVoiceInputEnable}
    >
      <SpeechToTextProvider
        onChangeProvider={onChangeAudioInputProvider}
        onChangeParameter={onChangeAudioInputParameter}
        provider={audioInputConfig.provider}
        parameters={audioInputConfig.parameters}
      />
      {audioInputConfig.provider && (
        <>
          <VADProvider
            className="m-0 mt-6"
            vadProvider={getParamValue('microphone.vad.provider', 'silero_vad')}
            onChangeVADProvider={v => {
              updateParameter('microphone.vad.provider', v);
            }}
            vadThreshold={getParamValue('microphone.vad.threshold', '0.8')}
            onChangeVadThreshold={(timeout: string) => {
              updateParameter('microphone.vad.threshold', timeout);
            }}
          />
          <NoiseCancellationProvider
            className="m-0 mt-6"
            noiseCancellationProvider={getParamValue(
              'microphone.denoising.provider',
              'rn_noise',
            )}
            onChangeNoiseCancellationProvider={v => {
              updateParameter('microphone.denoising.provider', v);
            }}
          />
          <EndOfSpeechProvider
            className="m-0 mt-6"
            endOfSpeechProvider={getParamValue(
              'microphone.eos.provider',
              'silence_based_eos',
            )}
            onChangeEndOfSpeechProvider={provider => {
              updateParameter('microphone.eos.provider', provider);
            }}
            endOfSepeechTimeout={getParamValue(
              'microphone.eos.timeout',
              '1000',
            )}
            onChangeEndOfSepeechTimeout={(timeout: string) => {
              updateParameter('microphone.eos.timeout', timeout);
            }}
          />
        </>
      )}
    </ConditionalInputGroup>
  );
};
