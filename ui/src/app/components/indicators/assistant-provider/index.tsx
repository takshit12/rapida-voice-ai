import { FC } from 'react';
import { Brain, Check, MinusCircle } from 'lucide-react';

export const AssistantProviderIndicator: FC<{
  provider: 'websocket' | 'agentkit' | 'provider-model';
  size?: 'small' | 'medium' | 'large';
}> = ({ provider, size = 'medium' }) => {
  const statusConfig = {
    WEBSOCKET: {
      bgColor: 'bg-gray-100 dark:bg-gray-800/50',
      textColor: 'text-gray-600 dark:text-gray-500',
      iconColor: 'dark:text-gray-400',
      ringColor: 'ring-gray-200 dark:ring-gray-800',
      Icon: Check,
      display: 'Websocket',
    },
    AGENTKIT: {
      bgColor: 'bg-gray-100 dark:bg-gray-800/50',
      textColor: 'text-gray-600 dark:text-gray-500',
      iconColor: 'dark:text-gray-400',
      ringColor: 'ring-gray-200 dark:ring-gray-800',
      Icon: MinusCircle,
      display: 'Agentkit',
    },
    PROVIDER_MODEL: {
      bgColor: 'bg-gray-100 dark:bg-gray-800/50',
      textColor: 'text-gray-600 dark:text-gray-500',
      iconColor: 'dark:text-gray-400',
      ringColor: 'ring-gray-200 dark:ring-gray-800',
      Icon: Brain,
      display: 'LLM',
    },
  };

  const config =
    statusConfig[provider] ||
    statusConfig[provider.toUpperCase()] ||
    statusConfig['PROVIDER_MODEL'];
  const { Icon } = config;

  // Size variants
  const sizeClasses = {
    small: {
      container: 'text-xs px-2 py-0.5 gap-1',
      icon: 12,
    },
    medium: {
      container: 'text-sm px-2.5 py-1 gap-1.5',
      icon: 16,
    },
    large: {
      container: 'text-base px-3 py-1.5 gap-2',
      icon: 18,
    },
  };

  const sizeClass = sizeClasses[size] || sizeClasses.medium;

  return (
    <span
      className={`shrink-0 gap-3 inline-flex items-center rounded-[2px] ${config.bgColor} ${config.textColor} font-medium ${sizeClass.container} ring-[0.5px] ring-inset ${config.ringColor}`}
    >
      <Icon size={sizeClass.icon} className={config.iconColor} />
      <span>{config.display}</span>
    </span>
  );
};
