import { Metadata, VaultCredential } from '@rapidaai/react';
import { Dropdown } from '@/app/components/dropdown';
import { ProviderComponentProps } from '@/app/components/providers';
import { ConfigureAnthropicTextProviderModel } from '@/app/components/providers/text/anthropic';
import {
  GetAnthropicTextProviderDefaultOptions,
  ValidateAnthropicTextProviderDefaultOptions,
} from '@/app/components/providers/text/anthropic/constants';
import {
  ConfigureAzureTextProviderModel,
  GetAzureTextProviderDefaultOptions,
  ValidateAzureTextProviderDefaultOptions,
} from '@/app/components/providers/text/azure';
import { ConfigureCohereTextProviderModel } from '@/app/components/providers/text/cohere';
import {
  GetCohereTextProviderDefaultOptions,
  ValidateCohereTextProviderDefaultOptions,
} from '@/app/components/providers/text/cohere/constants';
import { ConfigureGeminiTextProviderModel } from '@/app/components/providers/text/gemini';
import {
  GetGeminiTextProviderDefaultOptions,
  ValidateGeminiTextProviderDefaultOptions,
} from '@/app/components/providers/text/gemini/constants';
import { ConfigureOpenaiTextProviderModel } from '@/app/components/providers/text/openai';
import {
  GetOpenaiTextProviderDefaultOptions,
  ValidateOpenaiTextProviderDefaultOptions,
} from '@/app/components/providers/text/openai/constants';
import { cn } from '@/utils';
import { FC, useCallback } from 'react';
import { FieldSet } from '@/app/components/form/fieldset';
import { FormLabel } from '@/app/components/form-label';
import { CredentialDropdown } from '@/app/components/dropdown/credential-dropdown';
import { TEXT_PROVIDERS } from '@/providers';

/**
 *
 * @param provider
 * @param parameters
 * @returns
 */
export const GetDefaultTextProviderConfigIfInvalid = (
  provider: string,
  parameters: Metadata[],
): Metadata[] => {
  switch (provider) {
    case 'openai':
      return GetOpenaiTextProviderDefaultOptions(parameters);
    case 'azure-openai':
    case 'azure':
      return GetAzureTextProviderDefaultOptions(parameters);
    case 'gemini':
      return GetGeminiTextProviderDefaultOptions(parameters);
    case 'anthropic':
      return GetAnthropicTextProviderDefaultOptions(parameters);
    case 'cohere':
      return GetCohereTextProviderDefaultOptions(parameters);
    default:
      return parameters;
  }
};

/**
 *
 * @param provider
 * @param parameters
 * @returns
 */
export const ValidateTextProviderDefaultOptions = (
  provider: string,
  parameters: Metadata[],
): string | undefined => {
  switch (provider) {
    case 'openai':
      return ValidateOpenaiTextProviderDefaultOptions(parameters);
    case 'azure-openai':
    case 'azure':
      return ValidateAzureTextProviderDefaultOptions(parameters);
    case 'gemini':
      return ValidateGeminiTextProviderDefaultOptions(parameters);
    case 'anthropic':
      return ValidateAnthropicTextProviderDefaultOptions(parameters);
    case 'cohere':
      return ValidateCohereTextProviderDefaultOptions(parameters);
    default:
      return 'Please select a valid model and provider.';
  }
};

/**
 *
 * @param param0
 * @returns
 */
const TextProviderConfigComponent: FC<ProviderComponentProps> = ({
  provider,
  parameters,
  onChangeParameter,
}) => {
  switch (provider) {
    case 'openai':
      return (
        <ConfigureOpenaiTextProviderModel
          parameters={parameters}
          onParameterChange={onChangeParameter}
        />
      );
    case 'azure':
    case 'azure-openai':
      return (
        <ConfigureAzureTextProviderModel
          parameters={parameters}
          onParameterChange={onChangeParameter}
        />
      );
    case 'gemini':
      return (
        <ConfigureGeminiTextProviderModel
          parameters={parameters}
          onParameterChange={onChangeParameter}
        />
      );
    case 'anthropic':
      return (
        <ConfigureAnthropicTextProviderModel
          parameters={parameters}
          onParameterChange={onChangeParameter}
        />
      );
    case 'cohere':
      return (
        <ConfigureCohereTextProviderModel
          parameters={parameters}
          onParameterChange={onChangeParameter}
        />
      );
    default:
      return null;
  }
};

/**
 *
 * @param param0
 * @returns
 */
export const TextProvider: React.FC<ProviderComponentProps> = props => {
  const { provider, parameters, onChangeProvider, onChangeParameter } = props;
  const getParamValue = useCallback(
    (key: string) => {
      return parameters?.find(p => p.getKey() === key)?.getValue() ?? '';
    },
    [JSON.stringify(parameters)],
  );

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
    onChangeParameter(updatedParams);
  };

  return (
    <>
      <FieldSet>
        <FormLabel>Provider Model</FormLabel>
        <div
          className={cn(
            'outline-solid outline-transparent',
            'focus-within:outline-blue-600 focus:outline-blue-600 -outline-offset-1',
            'border-b border-gray-300 dark:border-gray-700',
            'dark:focus-within:border-blue-600 focus-within:border-blue-600',
            'transition-all duration-200 ease-in-out',
            'flex relative',
            'bg-light-background dark:bg-gray-950',
            'divide-x',
            'pt-px pl-px',
          )}
        >
          <div className="w-44 relative">
            <Dropdown
              className="max-w-full focus-within:border-none! outline-none! border-none! outline-hidden"
              currentValue={TEXT_PROVIDERS.find(x => x.code === provider)}
              setValue={v => {
                onChangeProvider(v.code);
              }}
              allValue={TEXT_PROVIDERS}
              placeholder="Select provider"
              option={c => {
                return (
                  <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
                    <img
                      alt=""
                      loading="lazy"
                      width={16}
                      height={16}
                      className="sm:h-4 sm:w-4 w-4 h-4 align-middle block shrink-0"
                      src={c.image}
                    />
                    <span className="truncate capitalize">{c.name}</span>
                  </span>
                );
              }}
              label={c => {
                return (
                  <span className="inline-flex items-center gap-2 sm:gap-2.5 max-w-full text-sm font-medium">
                    <img
                      alt=""
                      loading="lazy"
                      width={16}
                      height={16}
                      className="sm:h-4 sm:w-4 w-4 h-4 align-middle block shrink-0"
                      src={c.image}
                    />
                    <span className="truncate capitalize">{c.name}</span>
                  </span>
                );
              }}
            />
          </div>
          {/*  */}
          <TextProviderConfigComponent {...props} />
        </div>
      </FieldSet>
      {provider && (
        <CredentialDropdown
          onChangeCredential={(c: VaultCredential) => {
            updateParameter('rapida.credential_id', c.getId());
          }}
          provider={provider}
          currentCredential={getParamValue('rapida.credential_id')}
        />
      )}
    </>
  );
};
