import { Metadata, VaultCredential } from '@rapidaai/react';
import { Dropdown } from '@/app/components/dropdown';
import { ProviderComponentProps } from '@/app/components/providers';
import { ConfigureCohereEmbeddingModel } from '@/app/components/providers/embedding/cohere';
import {
  GetCohereEmbeddingDefaultOptions,
  ValidateCohereEmbeddingDefaultOptions,
} from '@/app/components/providers/embedding/cohere/constants';
import { ConfigureGoogleEmbeddingModel } from '@/app/components/providers/embedding/google';
import {
  GetGoogleEmbeddingDefaultOptions,
  ValidateGoogleEmbeddingDefaultOptions,
} from '@/app/components/providers/embedding/google/constants';
import { ConfigureOpenaiEmbeddingModel } from '@/app/components/providers/embedding/openai';
import {
  GetOpenaiEmbeddingDefaultOptions,
  ValidateOpenaiEmbeddingDefaultOptions,
} from '@/app/components/providers/embedding/openai/constants';
import { ConfigureVoyageEmbeddingModel } from '@/app/components/providers/embedding/voyageai';
import {
  GetVoyageEmbeddingDefaultOptions,
  ValidateVoyageEmbeddingDefaultOptions,
} from '@/app/components/providers/embedding/voyageai/constants';
import { cn } from '@/utils';
import { FC, useCallback } from 'react';
import { CredentialDropdown } from '@/app/components/dropdown/credential-dropdown';
import { FieldSet } from '@/app/components/form/fieldset';
import { FormLabel } from '@/app/components/form-label';
import { EMBEDDING_PROVIDERS } from '@/providers';

/**
 *
 * @param provider
 * @param parameters
 * @returns
 */
export const GetDefaultEmbeddingConfigIfInvalid = (
  provider: string,
  parameters: Metadata[],
): Metadata[] => {
  switch (provider) {
    case 'cohere':
      return GetCohereEmbeddingDefaultOptions(parameters);
    case 'openai':
      return GetOpenaiEmbeddingDefaultOptions(parameters);
    case 'google':
      return GetGoogleEmbeddingDefaultOptions(parameters);
    case 'voyageai':
      return GetVoyageEmbeddingDefaultOptions(parameters);
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
export const ValidateEmbeddingDefaultOptions = (
  provider: string,
  parameters: Metadata[],
): string | undefined => {
  switch (provider) {
    case 'cohere':
      return ValidateCohereEmbeddingDefaultOptions(parameters);
    case 'openai':
      return ValidateOpenaiEmbeddingDefaultOptions(parameters);
    case 'google':
      return ValidateGoogleEmbeddingDefaultOptions(parameters);
    case 'voyageai':
      return ValidateVoyageEmbeddingDefaultOptions(parameters);
    default:
      return 'Please select a valid provider and model for embedding';
  }
};

/**
 *
 * @param param0
 * @returns
 */
export const EmbeddingConfigComponent: FC<ProviderComponentProps> = ({
  provider,
  parameters,
  onChangeParameter,
}) => {
  switch (provider) {
    case 'cohere':
      return (
        <ConfigureCohereEmbeddingModel
          parameters={parameters}
          onParameterChange={onChangeParameter}
        />
      );
    case 'openai':
      return (
        <ConfigureOpenaiEmbeddingModel
          parameters={parameters}
          onParameterChange={onChangeParameter}
        />
      );
    case 'voyageai':
      return (
        <ConfigureVoyageEmbeddingModel
          parameters={parameters}
          onParameterChange={onChangeParameter}
        />
      );
    case 'gemini':
      return (
        <ConfigureGoogleEmbeddingModel
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
 * @param props
 * @returns
 */
export const EmbeddingProvider: React.FC<ProviderComponentProps> = props => {
  /**
   * all the parameters
   */
  const { provider, parameters, onChangeProvider, onChangeParameter } = props;
  /**
   * getter from paramerters
   */
  const getParamValue = useCallback(
    (key: string) => {
      return parameters?.find(p => p.getKey() === key)?.getValue() ?? '';
    },
    [JSON.stringify(parameters)],
  );

  /**
   *
   * @param key
   * @param value
   */
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

  /**
   *
   */
  return (
    <>
      <FieldSet>
        <FormLabel>Provider Model</FormLabel>
        <div
          className={cn(
            'p-px',
            'outline-solid outline-transparent',
            'focus-within:outline-blue-600 focus:outline-blue-600 -outline-offset-1',
            'border-b border-gray-300 dark:border-gray-700',
            'dark:focus-within:border-blue-600 focus-within:border-blue-600',
            'transition-all duration-200 ease-in-out',
            'bg-light-background dark:bg-gray-950',
            'divide-x',
            'flex relative',
          )}
        >
          <div className="w-44 relative">
            <Dropdown
              className="max-w-full focus-within:border-none! focus-within:outline-hidden! border-none! outline-hidden"
              currentValue={EMBEDDING_PROVIDERS.find(x => x.code === provider)}
              setValue={v => {
                onChangeProvider(v.code);
              }}
              allValue={EMBEDDING_PROVIDERS}
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
          <EmbeddingConfigComponent {...props} />
        </div>
      </FieldSet>
      {props && (
        <CredentialDropdown
          className="bg-white"
          onChangeCredential={(c: VaultCredential) => {
            updateParameter('rapida.credential_id', c.getId());
          }}
          currentCredential={getParamValue('rapida.credential_id')}
          provider={provider}
        />
      )}
    </>
  );
};
