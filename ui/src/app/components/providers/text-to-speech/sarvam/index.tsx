import { Metadata } from '@rapidaai/react';
import { Dropdown } from '@/app/components/dropdown';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import {
  SARVAM_LANGUAGE,
  SARVAM_TEXT_TO_SPEECH_MODEL,
  SARVAM_VOICE,
} from '@/providers';
import { useState } from 'react';
import { CustomValueDropdown } from '@/app/components/dropdown/custom-value-dropdown';
export { GetSarvamDefaultOptions, ValidateSarvamOptions } from './constant';

const renderOption = c => (
  <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
    {c.icon}
    <span className="truncate capitalize">{c.name}</span>
  </span>
);

export const ConfigureSarvamTextToSpeech: React.FC<{
  onParameterChange: (parameters: Metadata[]) => void;
  parameters: Metadata[] | null;
}> = ({ onParameterChange, parameters }) => {
  /**
   *
   */
  const [filteredVoices, setFilteredVoices] = useState(SARVAM_VOICE());

  /**
   *
   * @param key
   * @returns
   */
  const getParamValue = (key: string) => {
    return parameters?.find(p => p.getKey() === key)?.getValue() ?? '';
  };

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
      <FieldSet className="col-span-1">
        <FormLabel>Model</FormLabel>
        <Dropdown
          className="bg-light-background max-w-full dark:bg-gray-950"
          currentValue={SARVAM_TEXT_TO_SPEECH_MODEL().find(
            x => x.model_id === getParamValue('speak.model'),
          )}
          setValue={v => {
            updateParameter('speak.model', v.model_id);
          }}
          allValue={SARVAM_TEXT_TO_SPEECH_MODEL()}
          placeholder={`Select model`}
          option={renderOption}
          label={renderOption}
        />
      </FieldSet>
      <FieldSet className="col-span-1">
        <FormLabel>Voice</FormLabel>
        <CustomValueDropdown
          searchable
          className="bg-light-background max-w-full dark:bg-gray-950"
          currentValue={SARVAM_VOICE().find(
            x => x.id === getParamValue('speak.voice.id'),
          )}
          setValue={(v: { id: string }) => {
            updateParameter('speak.voice.id', v.id);
          }}
          allValue={filteredVoices}
          placeholder={`Select voice`}
          option={renderOption}
          label={renderOption}
          customValue
          onSearching={t => {
            const voices = SARVAM_VOICE();
            const v = t.target.value;
            if (v.length > 0) {
              setFilteredVoices(
                voices.filter(
                  voice =>
                    voice.name.toLowerCase().includes(v.toLowerCase()) ||
                    voice.id.toLowerCase().includes(v.toLowerCase()) ||
                    voice.language?.toLowerCase().includes(v.toLowerCase()),
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
      </FieldSet>

      <FieldSet className="col-span-1">
        <FormLabel>Language</FormLabel>
        <Dropdown
          className="bg-light-background max-w-full dark:bg-gray-950"
          currentValue={SARVAM_LANGUAGE().find(
            x => x.language_id === getParamValue('speak.language'),
          )}
          setValue={v => {
            updateParameter('speak.language', v.language_id);
          }}
          allValue={SARVAM_LANGUAGE()}
          placeholder={`Select languages`}
          option={renderOption}
          label={renderOption}
        />
      </FieldSet>
    </>
  );
};
