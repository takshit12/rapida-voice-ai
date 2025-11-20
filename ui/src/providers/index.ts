import productionProvider from './provider.production.json';
import developmentProvider from './provider.development.json';

export interface IntegrationProvider extends RapidaProvider {}
interface EndOfSpeechProvider extends RapidaProvider {}
interface NoiseCancellationProvider extends RapidaProvider {}
export interface RapidaProvider {
  code: string;
  name: string;
  featureList: string[];
  description?: string;
  image?: string;
  url?: string;
  configurations?: {
    name: string;
    type: string;
    label: string;
  }[];
}

export const allProvider = (): RapidaProvider[] => {
  const env = process.env.NODE_ENV || 'development';
  return env === 'production' ? productionProvider : developmentProvider;
};

export const EndOfSpeech = (): EndOfSpeechProvider[] => {
  return allProvider().filter(x => x.featureList.includes('end_of_speech'));
};

export const NoiseCancellation = (): NoiseCancellationProvider[] => {
  return allProvider().filter(x =>
    x.featureList.includes('noise_cancellation'),
  );
};

export const TEXT_PROVIDERS = allProvider().filter(x =>
  x.featureList.includes('text'),
);
export const INTEGRATION_PROVIDER: IntegrationProvider[] = allProvider().filter(
  x => x.featureList.includes('external'),
);
export const SPEECH_TO_TEXT_PROVIDER = allProvider().filter(x =>
  x.featureList.includes('stt'),
);
export const TEXT_TO_SPEECH_PROVIDER = allProvider().filter(x =>
  x.featureList.includes('tts'),
);
export const TEXT_TO_SPEECH = (name: string) =>
  allProvider()
    .filter(x => x.featureList.includes('tts'))
    .findLast(x => x.code === name);

export const STORAGE_PROVIDER = allProvider().filter(x =>
  x.featureList.includes('storage'),
);

export const RERANKER_PROVIDER = allProvider().filter(x =>
  x.featureList.includes('reranker'),
);

export const TELEPHONY_PROVIDER = allProvider().filter(x =>
  x.featureList.includes('telephony'),
);

export const EMBEDDING_PROVIDERS = allProvider().filter(x =>
  x.featureList.includes('embedding'),
);

/**
 *
 * @returns
 */

export const AZURE_VOICE = () => {
  return require('./azure/voices.json');
};

export const AZURE_LANGUAGE = () => {
  return require('./azure/languages.json');
};

/**
 *
 * @returns
 */
export const GOOGLE_CLOUD_VOICE = () => {
  return require('./google/interim.voice.json');
};

/**
 *
 * @returns
 */
export const DEEPGRAM_VOICE = () => {
  return require('./deepgram/voices.json');
};

/**
 * ElevEnlabs constants
 * @returns
 */
export const ELEVENLABS_MODEL = () => {
  return require('./elevenlabs/models.json');
};

export const ELEVENLABS_VOICE = () => {
  return require('./elevenlabs/voices.json');
};

export const ELEVENLABS_LANGUAGE = () => {
  return require('./elevenlabs/languages.json');
};

/**
 * cartesia
 */

//

export const CARTESIA_VOICE = () => {
  return require('./cartesia/voices.json');
};

export const CARTESIA_MODEL = () => {
  return require('./cartesia/models.json');
};

export const CARTESIA_LANGUAGE = () => {
  return require('./cartesia/languages.json');
};

export const CARTESIA_SPEED_OPTION = () => {
  return [
    { id: '', name: '' },
    { id: 'slowest', name: 'Slowest' },
    { id: 'slow', name: 'Slow' },
    { id: 'normal', name: 'Normal' },
    { id: 'fast', name: 'Fast' },
    { id: 'fastest', name: 'Fastest' },
  ];
};

export const CARTESIA_EMOTION_LEVEL_COMBINATION = [
  ...[
    { id: 'anger', name: 'Anger' },
    { id: 'positivity', name: 'Positivity' },
    { id: 'surprise', name: 'Surprise' },
    { id: 'sadness', name: 'Sadness' },
    { id: 'curiosity', name: 'Curiosity' },
  ].flatMap(m => m.id),
  ...[
    { id: 'anger', name: 'Anger' },
    { id: 'positivity', name: 'Positivity' },
    { id: 'surprise', name: 'Surprise' },
    { id: 'sadness', name: 'Sadness' },
    { id: 'curiosity', name: 'Curiosity' },
  ].flatMap(emotion =>
    [
      { id: 'lowest', name: 'Lowest' },
      { id: 'low', name: 'Low' },
      { id: 'high', name: 'High' },
      { id: 'highest', name: 'Highest' },
    ].map(level => `${emotion.id}:${level.id}`),
  ),
];

/**
 *
 */

export const GEMINI_MODEL = () => {
  return require('./gemini/models.json');
};
