import { Metadata } from '@rapidaai/react';
import { Dropdown } from '@/app/components/dropdown';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import {
  CARTESIA_EMOTION_LEVEL_COMBINATION,
  CARTESIA_LANGUAGE,
  CARTESIA_MODEL,
  CARTESIA_SPEED_OPTION,
  CARTESIA_VOICE,
} from '@/providers';
import { ILinkBorderButton } from '@/app/components/form/button';
import { useState } from 'react';
import { CustomValueDropdown } from '@/app/components/dropdown/custom-value-dropdown';
import { ExternalLink } from 'lucide-react';
export { GetCartesiaDefaultOptions, ValidateCartesiaOptions } from './constant';

const renderOption = c => (
  <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
    {c.icon}
    <span className="truncate capitalize">{c.name}</span>
  </span>
);

export const ConfigureCartesiaTextToSpeech: React.FC<{
  onParameterChange: (parameters: Metadata[]) => void;
  parameters: Metadata[] | null;
}> = ({ onParameterChange, parameters }) => {
  /**
   *
   */
  const [filteredVoices, setFilteredVoices] = useState(CARTESIA_VOICE());

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
          currentValue={CARTESIA_MODEL().find(
            x => x.id === getParamValue('speak.model'),
          )}
          setValue={v => {
            updateParameter('speak.model', v.id);
          }}
          allValue={CARTESIA_MODEL()}
          placeholder={`Select model`}
          option={renderOption}
          label={renderOption}
        />
      </FieldSet>
      <FieldSet className="col-span-2">
        <FormLabel>Voice</FormLabel>
        <div className="flex">
          <CustomValueDropdown
            searchable
            className="bg-light-background max-w-full dark:bg-gray-950"
            currentValue={CARTESIA_VOICE().find(
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
              const voices = CARTESIA_VOICE();
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
          <ILinkBorderButton
            target="_blank"
            href={`/integration/models/cartesia?query=${getParamValue('speak.voice.id')}`}
            className="h-10 text-sm p-2 px-3 bg-light-background max-w-full dark:bg-gray-950 border-b"
          >
            <ExternalLink className="w-4 h-4" strokeWidth={1.5} />
          </ILinkBorderButton>
        </div>
      </FieldSet>

      <FieldSet className="col-span-1">
        <FormLabel>Language</FormLabel>
        <Dropdown
          className="bg-light-background max-w-full dark:bg-gray-950"
          currentValue={CARTESIA_LANGUAGE().find(
            x => x.code === getParamValue('speak.language'),
          )}
          setValue={v => {
            updateParameter('speak.language', v.code);
          }}
          allValue={CARTESIA_LANGUAGE()}
          placeholder={`Select languages`}
          option={renderOption}
          label={renderOption}
        />
      </FieldSet>
      <FieldSet className="col-span-1">
        <FormLabel>Speed (Experimental)</FormLabel>
        <Dropdown
          className="bg-light-background max-w-full dark:bg-gray-950"
          currentValue={CARTESIA_SPEED_OPTION().find(
            x =>
              x.id ===
              getParamValue('speak.voice.__experimental_controls.speed'),
          )}
          setValue={v => {
            updateParameter('speak.voice.__experimental_controls.speed', v.id);
          }}
          allValue={CARTESIA_SPEED_OPTION()}
          placeholder={`Select Speed`}
          option={renderOption}
          label={renderOption}
        />
      </FieldSet>

      <FieldSet className="relative col-span-2">
        <FormLabel>Emotion (Experimental)</FormLabel>
        <Dropdown
          multiple
          className="bg-light-background dark:bg-gray-950 max-w-6xl"
          currentValue={getParamValue(
            'speak.voice.__experimental_controls.emotion',
          ).split('<|||>')}
          setValue={v => {
            updateParameter(
              'speak.voice.__experimental_controls.emotion',
              v.join('<|||>'),
            );
          }}
          allValue={CARTESIA_EMOTION_LEVEL_COMBINATION}
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
      </FieldSet>
    </>
  );
};
