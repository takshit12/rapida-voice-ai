import {
  AssistantWebpluginDeployment,
  DeploymentAudioProvider,
} from '@rapidaai/react';
import { Tab } from '@/app/components/tab';
import { cn } from '@/utils';
import { ChevronRight } from 'lucide-react';
import { ModalProps } from '@/app/components/base/modal';
import { RightSideModal } from '@/app/components/base/modal/right-side-modal';
import { FieldSet } from '@/app/components/form/fieldset';
import { FormLabel } from '@/app/components/form-label';
import { CONFIG } from '@/configs';
import { CopyButton } from '@/app/components/form/button/copy-button';
import { InputHelper } from '@/app/components/input-helper';
import { YellowNoticeBlock } from '@/app/components/container/message/notice-block';
import { ProviderPill } from '@/app/components/pill/provider-model-pill';
import { FC } from 'react';
import { CodeHighlighting } from '@/app/components/code-highlighting';

interface AssistantWebWidgetlDeploymentDialogProps extends ModalProps {
  deployment: AssistantWebpluginDeployment;
}
/**
 *
 * @param props
 * @returns
 */
export function AssistantWebWidgetlDeploymentDialog(
  props: AssistantWebWidgetlDeploymentDialogProps,
) {
  return (
    <RightSideModal
      modalOpen={props.modalOpen}
      setModalOpen={props.setModalOpen}
      className="w-2/3 xl:w-1/3 flex-1"
    >
      <div className="flex items-center p-4 border-b text-base/6 font-medium">
        <div className="font-medium">Assistant</div>
        <ChevronRight size={18} className="mx-2" />
        <div className="font-medium">Deployment</div>
        <ChevronRight size={18} className="mx-2" />
        <div className="font-medium">vrsn_dpl_{props.deployment.getId()}</div>
      </div>
      <div className="relative overflow-auto h-[calc(100vh-50px)] flex flex-col flex-1">
        <Tab
          active="Integration"
          className={cn(
            'text-sm',
            'bg-gray-50 border-b dark:bg-gray-900 dark:border-gray-800 sticky top-0 z-1',
          )}
          tabs={[
            {
              label: 'Integration',
              element: (
                <div className="flex-1 px-4 space-y-8">
                  <FieldSet>
                    <div className="font-medium border-b -mx-4 px-4 py-2">
                      Add script to your website
                    </div>
                    <CodeHighlighting
                      className="min-h-[20px]"
                      code='<script src="https://cdn-01.rapida.ai/public/scripts/app.min.js" defer></script>'
                    ></CodeHighlighting>
                  </FieldSet>
                  <FieldSet>
                    <div className="font-medium border-b -mx-4 px-4 py-2">
                      Initialize assistant
                    </div>
                    <CodeHighlighting
                      className="min-h-[200px]"
                      code={`
                        <script>window.chatbotConfig = {
theme: {
    color: "black",
},
assistant_id: "2139456643765633024",
token:
    "",
user: {
    name: "Guest",
}
}</script>`.trim()}
                    ></CodeHighlighting>
                  </FieldSet>
                </div>
              ),
            },
            {
              label: 'Audio',
              element: (
                <div className="flex-1 space-y-8">
                  <VoiceInput deployment={props.deployment?.getInputaudio()} />

                  <VoiceOutput
                    deployment={props.deployment?.getOutputaudio()}
                  />
                </div>
              ),
            },
          ]}
        />
      </div>
    </RightSideModal>
  );
}

const VoiceInput: FC<{ deployment?: DeploymentAudioProvider }> = ({
  deployment,
}) => (
  <div className="">
    <div className="flex items-center space-x-2 border-b py-1 px-4 h-10">
      <h4 className="font-medium">Speech to text</h4>
    </div>
    {deployment?.getAudiooptionsList() ? (
      deployment?.getAudiooptionsList().length > 0 && (
        <div className="text-xs text-gray-500 dark:text-gray-400 py-3 px-3 space-y-6">
          <FieldSet>
            <FormLabel>Provider</FormLabel>
            <ProviderPill provider={deployment?.getAudioprovider()} />
          </FieldSet>
          <div className="grid grid-cols-1 gap-4">
            {deployment
              ?.getAudiooptionsList()
              .filter(d => d.getValue())
              .filter(d => d.getKey().startsWith('listen.'))
              .map((detail, index) => (
                <FieldSet key={index}>
                  <FormLabel>{detail.getKey()}</FormLabel>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 dark:bg-gray-950 bg-gray-100 px-3 py-2 font-mono text-xs min-w-0 overflow-hidden">
                      {detail.getValue()}
                    </code>
                    <div className="flex shrink-0 border divide-x">
                      <CopyButton className="h-7 w-7">
                        {detail.getValue()}
                      </CopyButton>
                    </div>
                  </div>
                </FieldSet>
              ))}
          </div>
        </div>
      )
    ) : (
      <YellowNoticeBlock>Voice input is not enabled</YellowNoticeBlock>
    )}
  </div>
);

const VoiceOutput: FC<{ deployment?: DeploymentAudioProvider }> = ({
  deployment,
}) => (
  <div>
    <div className="flex items-center space-x-2 border-b py-2 px-4  h-10">
      <h4 className="font-medium">Text to speech</h4>
    </div>
    {deployment?.getAudiooptionsList() ? (
      deployment?.getAudiooptionsList().length > 0 && (
        <div className="text-xs text-gray-500 dark:text-gray-400 py-3 px-3 space-y-6">
          <FieldSet>
            <FormLabel>Provider</FormLabel>
            <ProviderPill provider={deployment?.getAudioprovider()} />
          </FieldSet>
          <div className="grid grid-cols-1 gap-4">
            {deployment
              ?.getAudiooptionsList()
              .filter(d => d.getValue())
              .filter(d => d.getKey().startsWith('speak.'))
              .map((detail, index) => (
                <FieldSet key={index}>
                  <FormLabel>{detail.getKey()}</FormLabel>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 dark:bg-gray-950 bg-gray-100 px-3 py-2 font-mono text-xs min-w-0 overflow-hidden">
                      {detail.getValue()}
                    </code>

                    <div className="flex shrink-0 border divide-x">
                      <CopyButton className="h-7 w-7">
                        {detail.getValue()}
                      </CopyButton>
                    </div>
                  </div>
                </FieldSet>
              ))}
          </div>
        </div>
      )
    ) : (
      <YellowNoticeBlock>Voice output is not enabled</YellowNoticeBlock>
    )}
  </div>
);
