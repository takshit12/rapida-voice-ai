import { Metadata } from '@rapidaai/react';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import { useMemo, useState } from 'react';
import {
  SMALLEST_LANGUAGE,
  SMALLEST_MODEL,
  SMALLEST_VOICE,
} from '@/providers';
import { CustomValueDropdown } from '@/app/components/dropdown/custom-value-dropdown';
import { Dropdown } from '@/app/components/dropdown';
export {
  GetSmallestDefaultOptions,
  ValidateSmallestOptions,
} from './constant';

const renderOption = (c: { name: string }) => {
  return (
    <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
      <span className="truncate capitalize">{c.name}</span>
    </span>
  );
};

export const ConfigureSmallestTextToSpeech: React.FC<{
  onParameterChange: (parameters: Metadata[]) => void;
  parameters: Metadata[] | null;
}> = ({ onParameterChange, parameters }) => {
  const getParamValue = (key: string) =>
    parameters?.find(p => p.getKey() === key)?.getValue() ?? '';

  const selectedModel = getParamValue('speak.model');

  // Filter voices by selected model
  const modelVoices = useMemo(() => {
    const voices = SMALLEST_VOICE();
    if (!selectedModel) return voices;
    return voices.filter(
      (v: { models?: string[] }) =>
        !v.models || v.models.includes(selectedModel),
    );
  }, [selectedModel]);

  const [filteredVoices, setFilteredVoices] = useState(modelVoices);

  // Filter languages by selected model
  const modelLanguages = useMemo(() => {
    const languages = SMALLEST_LANGUAGE();
    if (!selectedModel) return languages;
    return languages.filter(
      (l: { models?: string[] }) =>
        !l.models || l.models.includes(selectedModel),
    );
  }, [selectedModel]);

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

  const handleModelChange = (v: { model_id: string }) => {
    updateParameter('speak.model', v.model_id);
    // Reset voice and language when model changes since available options differ
    updateParameter('speak.voice.id', '');
    updateParameter('speak.language', '');
    const voices = SMALLEST_VOICE().filter(
      (voice: { models?: string[] }) =>
        !voice.models || voice.models.includes(v.model_id),
    );
    setFilteredVoices(voices);
  };

  return (
    <>
      <FieldSet className="col-span-1" key="speak.model">
        <FormLabel>Models</FormLabel>
        <Dropdown
          className="bg-light-background max-w-full dark:bg-gray-950"
          currentValue={SMALLEST_MODEL().find(
            x => x.model_id === getParamValue('speak.model'),
          )}
          setValue={handleModelChange}
          allValue={SMALLEST_MODEL()}
          placeholder="Select model"
          option={renderOption}
          label={renderOption}
        />
      </FieldSet>
      <FieldSet className="col-span-1" key="speak.voice.id">
        <FormLabel>Voice</FormLabel>
        <CustomValueDropdown
          searchable
          className="bg-light-background max-w-full dark:bg-gray-950"
          currentValue={modelVoices.find(
            x => x.voice_id === getParamValue('speak.voice.id'),
          )}
          setValue={(v: { voice_id: string }) =>
            updateParameter('speak.voice.id', v.voice_id)
          }
          allValue={filteredVoices}
          customValue
          onSearching={t => {
            const v = t.target.value;
            if (v.length > 0) {
              setFilteredVoices(
                modelVoices.filter(
                  voice =>
                    voice.name.toLowerCase().includes(v.toLowerCase()) ||
                    voice.voice_id.toLowerCase().includes(v.toLowerCase()) ||
                    voice.tags?.language
                      ?.toLowerCase()
                      .includes(v.toLowerCase()) ||
                    voice.tags?.gender
                      ?.toLowerCase()
                      .includes(v.toLowerCase()),
                ),
              );
              return;
            }
            setFilteredVoices(modelVoices);
          }}
          onAddCustomValue={vl => {
            filteredVoices.push({ voice_id: vl, name: vl });
            updateParameter('speak.voice.id', vl);
          }}
          placeholder="Select voice"
          option={renderOption}
          label={renderOption}
        />
      </FieldSet>

      <FieldSet className="col-span-1" key="speak.language">
        <FormLabel>Language</FormLabel>
        <Dropdown
          className="bg-light-background max-w-full dark:bg-gray-950"
          currentValue={modelLanguages.find(
            x => x.language_id === getParamValue('speak.language'),
          )}
          setValue={(v: { language_id: string }) =>
            updateParameter('speak.language', v.language_id)
          }
          allValue={modelLanguages}
          placeholder="Select language"
          option={renderOption}
          label={renderOption}
        />
      </FieldSet>
    </>
  );
};
