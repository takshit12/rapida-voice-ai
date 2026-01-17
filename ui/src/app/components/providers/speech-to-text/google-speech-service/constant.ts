
import { SetMetadata } from '@/utils/metadata';
import { Metadata } from '@rapidaai/react';

export const GetGoogleDefaultOptions = (current: Metadata[]): Metadata[] => {
  const mtds: Metadata[] = [];
  // Define the keys we want to keep
  const keysToKeep = [
    'rapida.credential_id',
    'listen.language',
    'listen.model',
    'listen.threshold',
    'listen.region',
  ];

  // Function to create or update metadata
  const addMetadata = (
    key: string,
    defaultValue?: string,
    validationFn?: (value: string) => boolean,
  ) => {
    const metadata = SetMetadata(current, key, defaultValue, validationFn);
    if (metadata) mtds.push(metadata);
  };

  addMetadata('rapida.credential_id');
  addMetadata('listen.language');
  addMetadata('listen.threshold', '0.5');
  addMetadata('listen.model');
  addMetadata('listen.region');

  // Only return metadata for the keys we want to keep
  return [
    ...mtds.filter(m => keysToKeep.includes(m.getKey())),
    ...current.filter(m => m.getKey().startsWith('microphone.')),
  ];
};

export const ValidateGoogleOptions = (
  options: Metadata[],
): string | undefined => {
  const credentialID = options.find(
    opt => opt.getKey() === 'rapida.credential_id',
  );
  if (
    !credentialID ||
    !credentialID.getValue() ||
    credentialID.getValue().length === 0
  ) {
    return 'Please provide a valid credential for google speech to text.';
  }

  return undefined;
};
