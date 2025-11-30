import { AssistantConversationMessage, toContentText } from '@rapidaai/react';
import { Metadata } from '@rapidaai/react';
import { Tab } from '@/app/components/tab';
import { cn } from '@/utils';
import { ChevronRight } from 'lucide-react';
import { ModalProps } from '@/app/components/base/modal';
import { RightSideModal } from '@/app/components/base/modal/right-side-modal';
import { CodeHighlighting } from '@/app/components/code-highlighting';
import { MarkdownViewer } from '@/app/components/markdown-viewer';
import {
  BlueNoticeBlock,
  YellowNoticeBlock,
} from '@/app/components/container/message/notice-block';
import { FC } from 'react';

interface ConversationLogDialogProps extends ModalProps {
  currentAssistantMessage: AssistantConversationMessage;
}
/**
 *
 * @param props
 * @returns
 */
export function ConversationLogDialog(props: ConversationLogDialogProps) {
  return (
    <RightSideModal
      modalOpen={props.modalOpen}
      setModalOpen={props.setModalOpen}
      className="w-2/3 xl:w-1/3 flex-1"
    >
      <div className="flex items-center p-4 border-b">
        <div className="font-medium text-lg">Log</div>
        <ChevronRight size={18} className="mx-2" />
        <div className="font-medium text-lg">Conversation</div>
        <ChevronRight size={18} className="mx-2" />
        <div className="font-medium text-base">
          {props.currentAssistantMessage.getAssistantconversationid()}
        </div>
      </div>
      <div className="relative overflow-auto h-[calc(100vh-50px)] flex flex-col flex-1">
        <Tab
          active="Request"
          className={cn(
            'text-sm',
            'bg-gray-50 border-b dark:bg-gray-900 dark:border-gray-800 sticky top-0 z-1',
          )}
          tabs={[
            {
              label: 'Request',
              element: (
                <div className="flex-1 p-4 space-y-8">
                  {props.currentAssistantMessage.getRequest() ? (
                    <div className="border rounded-[2px]">
                      <MarkdownViewer
                        text={toContentText(
                          props.currentAssistantMessage
                            .getRequest()
                            ?.getContentsList(),
                        )}
                      />
                    </div>
                  ) : (
                    <YellowNoticeBlock>
                      Request will be available here after the completion of
                      execution
                    </YellowNoticeBlock>
                  )}
                </div>
              ),
            },
            {
              label: 'Response',
              element: (
                <div className="flex-1 p-4 space-y-8">
                  {props.currentAssistantMessage.getResponse() ? (
                    <div className="border rounded-[2px]">
                      <MarkdownViewer
                        text={toContentText(
                          props.currentAssistantMessage
                            .getResponse()
                            ?.getContentsList(),
                        )}
                      />
                    </div>
                  ) : (
                    <YellowNoticeBlock>
                      Response will be available here after the completion of
                      execution.
                    </YellowNoticeBlock>
                  )}
                </div>
              ),
            },
            {
              label: 'Metrics',
              element: (
                <CodeHighlighting
                  lang="json"
                  lineNumbers={false}
                  foldGutter={false}
                  code={JSON.stringify(
                    props.currentAssistantMessage
                      ?.getMetricsList()
                      .map(metric => metric.toObject()),
                    null,
                    2,
                  )}
                />
              ),
            },
            {
              label: 'Metadata',
              element: (
                <MessageMetadatas
                  metadata={props.currentAssistantMessage.getMetadataList()}
                />
              ),
            },
          ]}
        />
      </div>
    </RightSideModal>
  );
}

const MessageMetadatas: FC<{ metadata: Array<Metadata> }> = ({ metadata }) => {
  if (metadata.length <= 0)
    return (
      <BlueNoticeBlock>There are no metdata for given message.</BlueNoticeBlock>
    );
  return (
    <div className="flex flex-col w-full">
      {metadata.map((x, idx) => {
        return (
          <div
            className="flex justify-between w-full items-center border-[0.5px] rounded-[2px]"
            key={`metadata-idx-${idx}`}
          >
            <div className="py-3 px-4 flex items-center gap-2">
              <span className="capitalize">{x.getKey()}</span>
            </div>
            <div className="py-3 px-4 ">
              <div className="flex items-center">{x.getValue()}</div>
            </div>
          </div>
        );
      })}
    </div>
  );
};
