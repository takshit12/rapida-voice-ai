import { Assistant } from '@rapidaai/react';
import { SectionLoader } from '@/app/components/loader/section-loader';
import { AssistantAnalytics } from '@/app/pages/assistant/view/overview/assistant-analytics';
import { useRapidaStore } from '@/hooks';
import { FC } from 'react';
import { YellowNoticeBlock } from '@/app/components/container/message/notice-block';
import { ExternalLink, Info } from 'lucide-react';
import { useGlobalNavigation } from '@/hooks/use-global-navigator';
/**
 *
 * @param props
 * @returns
 */
export const Overview: FC<{ currentAssistant: Assistant }> = (props: {
  currentAssistant: Assistant;
}) => {
  const rapidaContext = useRapidaStore();
  const navigation = useGlobalNavigation();

  if (rapidaContext.loading) {
    return (
      <div className="h-full flex flex-col items-center justify-center">
        <SectionLoader />
      </div>
    );
  }

  return (
    <div className="flex flex-col flex-1 grow">
      {!props.currentAssistant.getApideployment() &&
        !props.currentAssistant.getDebuggerdeployment() &&
        !props.currentAssistant.getWebplugindeployment() &&
        !props.currentAssistant.getPhonedeployment() && (
          <YellowNoticeBlock className="flex items-center">
            <Info className="shrink-0 w-4 h-4" strokeWidth={1.5} />
            <div className="ms-3 text-sm font-medium">
              <strong className="font-semibold">
                Your assistant is ready, but not live yet,
              </strong>{' '}
              It looks like your assistant isn’t deployed to any channel.
            </div>
            <button
              type="button"
              onClick={() => {
                navigation.goToDeploymentAssistant(
                  props.currentAssistant.getId(),
                );
              }}
              className="h-7 flex items-center font-medium hover:underline ml-auto text-yellow-600"
            >
              Enable deployment
              <ExternalLink
                className="shrink-0 w-4 h-4 ml-1.5"
                strokeWidth={1.5}
              />
            </button>
          </YellowNoticeBlock>
        )}
      <AssistantAnalytics assistant={props.currentAssistant} />
    </div>
  );
};
