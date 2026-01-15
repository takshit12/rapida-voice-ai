import { Metadata } from '@rapidaai/react';
import { Dropdown } from '@/app/components/dropdown';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import { CustomValueDropdown } from '@/app/components/dropdown/custom-value-dropdown';
import {
  AZURE_TEXT_TO_SPEECH_LANGUAGE,
  AZURE_TEXT_TO_SPEECH_VOICE,
} from '@/providers';
import { useState } from 'react';
import { ILinkBorderButton } from '@/app/components/form/button';
import { ExternalLink } from 'lucide-react';

export { GetAzureDefaultOptions, ValidateAzureOptions } from './constant';
const renderVoiceOption = (c: { icon: React.ReactNode; shortName: string }) => (
  <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
    {c.icon}
    <span className="truncate capitalize">{c.shortName}</span>
  </span>
);

const renderLanguageOption = (c: { icon: React.ReactNode; name: string }) => (
  <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
    {c.icon}
    <span className="truncate capitalize">{c.name}</span>
  </span>
);

export const ConfigureAzureTextToSpeech: React.FC<{
  onParameterChange: (parameters: Metadata[]) => void;
  parameters: Metadata[] | null;
}> = ({ onParameterChange, parameters }) => {
  /**
   *
   */
  const [filteredVoices, setFilteredVoices] = useState(
    AZURE_TEXT_TO_SPEECH_VOICE(),
  );

  /**
   *
   */
  const [filterLanguages, setFilterLanguages] = useState(
    AZURE_TEXT_TO_SPEECH_LANGUAGE(),
  );
  /**
   *
   * @param key
   * @returns
   */
  const getParamValue = (key: string) =>
    parameters?.find(p => p.getKey() === key)?.getValue() ?? '';

  //
  const updateParameter = (key: string, value: string) => {
    const updatedParams = [...(parameters || [])];
    const existingIndex = updatedParams.findIndex(p => p.getKey() === key);
    const newParam = new Metadata();
    newParam.setKey(key);
    newParam.setValue(value);
    if (existingIndex >= 0) {
      updatedParams[existingIndex] = newParam;
    } else {
      updatedParams.push(newParam);
    }
    onParameterChange(updatedParams);
  };

  return (
    <>
      <FieldSet className="col-span-2">
        <FormLabel>Voice</FormLabel>
        <div className="flex">
          <CustomValueDropdown
            searchable
            className="bg-light-background max-w-full dark:bg-gray-950"
            currentValue={AZURE_TEXT_TO_SPEECH_VOICE().find(
              x => x.shortName === getParamValue('speak.voice.id'),
            )}
            setValue={(v: { shortName: string }) => {
              updateParameter('speak.voice.id', v.shortName);
            }}
            allValue={filteredVoices}
            placeholder={`Select voice`}
            option={renderVoiceOption}
            label={renderVoiceOption}
            customValue
            onSearching={t => {
              const voices = AZURE_TEXT_TO_SPEECH_VOICE();
              const v = t.target.value;
              if (v.length > 0) {
                setFilteredVoices(
                  voices.filter(
                    voice =>
                      voice.properties.DisplayName.toLowerCase().includes(
                        v.toLowerCase(),
                      ) ||
                      voice.shortName.toLowerCase().includes(v.toLowerCase()) ||
                      voice.locale?.toLowerCase().includes(v.toLowerCase()),
                  ),
                );
                return;
              }
              setFilteredVoices(voices);
            }}
            onAddCustomValue={vl => {
              updateParameter('speak.voice.id', vl);
            }}
          />
          <ILinkBorderButton
            target="_blank"
            href={`/integration/models/azure-speech-service?query=${getParamValue('speak.voice.id')}`}
            className="h-10 text-sm p-2 px-3 bg-light-background max-w-full dark:bg-gray-950 border-b"
          >
            <ExternalLink className="w-4 h-4" strokeWidth={1.5} />
          </ILinkBorderButton>
        </div>
      </FieldSet>

      <FieldSet className="col-span-1">
        <FormLabel>Language</FormLabel>
        <Dropdown
          searchable
          className="bg-light-background max-w-full dark:bg-gray-950"
          currentValue={AZURE_TEXT_TO_SPEECH_LANGUAGE().find(
            x => x.code === getParamValue('speak.language'),
          )}
          setValue={v => {
            updateParameter('speak.language', v.code);
          }}
          allValue={filterLanguages}
          placeholder={`Select language`}
          option={renderLanguageOption}
          label={renderLanguageOption}
          onSearching={t => {
            const lanaguages = AZURE_TEXT_TO_SPEECH_LANGUAGE();
            const v = t.target.value;
            if (v.length > 0) {
              setFilterLanguages(
                lanaguages.filter(
                  lg =>
                    lg.name.toLowerCase().includes(v.toLowerCase()) ||
                    lg.code?.toLowerCase().includes(v.toLowerCase()),
                ),
              );
              return;
            }
            setFilterLanguages(lanaguages);
          }}
        />
      </FieldSet>
    </>
  );
};
