import { IBlueBorderButton } from '@/app/components/form/button';
import { useAssistantProviderPageStore } from '@/hooks';
import { useRapidaStore } from '@/hooks';
import { useCredential } from '@/hooks/use-credential';
import { useEffect } from 'react';
import toast from 'react-hot-toast/headless';
import { cn } from '@/utils';
import { Assistant, GetAllAssistantProviderResponse } from '@rapidaai/react';
import { RevisionIndicator } from '@/app/components/indicators/revision';
import { SectionLoader } from '@/app/components/loader/section-loader';
import { CircleSlash } from 'lucide-react';
import { TableSection } from '@/app/components/sections/table-section';
import { ScrollableResizableTable } from '@/app/components/data-table';
import { TableRow } from '@/app/components/base/tables/table-row';
import { TableCell } from '@/app/components/base/tables/table-cell';
import { CopyCell } from '@/app/components/base/tables/copy-cell';
import { AssistantProviderIndicator } from '@/app/components/indicators/assistant-provider';
import { DateCell } from '@/app/components/base/tables/date-cell';
import { NameCell } from '@/app/components/base/tables/name-cell';

interface VersionProps {
  assistant: Assistant;
  onReload: () => void;
}

export function Version(props: VersionProps) {
  const [userId, token, projectId] = useCredential();
  const rapidaContext = useRapidaStore();
  const assistantProviderAction = useAssistantProviderPageStore();

  useEffect(() => {
    rapidaContext.showLoader();
    assistantProviderAction.onChangeAssistant(props.assistant);
    assistantProviderAction.getAssistantProviders(
      props.assistant.getId(),
      projectId,
      token,
      userId,
      (err: string) => {
        rapidaContext.hideLoader();
        toast.error(err);
      },
      data => {
        rapidaContext.hideLoader();
      },
    );
  }, [
    props.assistant.getId(),
    projectId,
    assistantProviderAction.page,
    assistantProviderAction.pageSize,
    assistantProviderAction.criteria,
  ]);

  const deployRevision = (
    assistantProvider: string,
    assistantProviderId: string,
  ) => {
    rapidaContext.showLoader('overlay');
    assistantProviderAction.onReleaseVersion(
      assistantProvider,
      assistantProviderId,
      projectId,
      token,
      userId,
      error => {
        rapidaContext.hideLoader();
        toast.error(error);
      },
      e => {
        toast.success(
          'New version of assistant has been deployed successfully.',
        );
        assistantProviderAction.onChangeAssistant(e);
        props.onReload();
        rapidaContext.hideLoader();
      },
    );
  };
  if (rapidaContext.loading) {
    return (
      <div className="h-full flex flex-col items-center justify-center">
        <SectionLoader />
      </div>
    );
  }

  return (
    <TableSection>
      <ScrollableResizableTable
        isActionable={false}
        clms={assistantProviderAction.columns.filter(x => x.visible)}
      >
        {assistantProviderAction.assistantProviders.map((apm, idx) => {
          switch (apm.getAssistantproviderCase()) {
            case GetAllAssistantProviderResponse.AssistantProvider
              .AssistantproviderCase.ASSISTANTPROVIDERMODEL:
              return (
                <TableRow key={idx} data-id={idx}>
                  <CopyCell>
                    {`vrsn_${apm.getAssistantprovidermodel()?.getId()}`}
                  </CopyCell>
                  <TableCell>
                    <AssistantProviderIndicator provider="provider-model" />
                  </TableCell>
                  <TableCell>
                    {apm.getAssistantprovidermodel()?.getDescription()
                      ? apm.getAssistantprovidermodel()?.getDescription()
                      : 'Initial assistant version'}
                  </TableCell>
                  <TableCell>
                    {assistantProviderAction.assistant?.getAssistantproviderid() !==
                    apm.getAssistantprovidermodel()?.getId() ? (
                      <IBlueBorderButton
                        disabled={
                          assistantProviderAction.assistant?.getAssistantproviderid() ===
                          apm.getAssistantprovidermodel()?.getId()
                        }
                        className={cn(
                          'shrink-0 !bg-blue-600/10 hover:!bg-blue-600 border-none h-8 ring-[0.5] ring-blue-600',
                        )}
                        onClick={() => {
                          deployRevision(
                            'MODEL',
                            apm.getAssistantprovidermodel()?.getId()!,
                          );
                        }}
                      >
                        <CircleSlash
                          className="w-3.5 h-3.5 mr-2"
                          strokeWidth={1.5}
                        />
                        <span>Deploy revision</span>
                      </IBlueBorderButton>
                    ) : (
                      <RevisionIndicator
                        status={
                          assistantProviderAction.assistant?.getAssistantproviderid() ===
                          apm.getAssistantprovidermodel()?.getId()
                            ? 'DEPLOYED'
                            : 'NOT_DEPLOYED'
                        }
                      />
                    )}
                  </TableCell>
                  <NameCell data-id={apm.getAssistantprovidermodel()?.getId()}>
                    {apm.getAssistantprovidermodel()?.getCreateduser() &&
                      apm
                        .getAssistantprovidermodel()
                        ?.getCreateduser()
                        ?.getName()!}
                  </NameCell>
                  <DateCell
                    date={apm.getAssistantprovidermodel()?.getCreateddate()}
                  />
                </TableRow>
              );
            case GetAllAssistantProviderResponse.AssistantProvider
              .AssistantproviderCase.ASSISTANTPROVIDERAGENTKIT:
              return (
                <TableRow key={idx} className="cursor-pointer" data-id={idx}>
                  <CopyCell>
                    {`vrsn_${apm.getAssistantprovideragentkit()?.getId()}`}
                  </CopyCell>
                  <TableCell>
                    <AssistantProviderIndicator provider="agentkit" />
                  </TableCell>
                  <TableCell>
                    {apm.getAssistantprovideragentkit()?.getDescription()
                      ? apm.getAssistantprovideragentkit()?.getDescription()
                      : 'Initial assistant version'}
                  </TableCell>
                  <TableCell>
                    {assistantProviderAction.assistant?.getAssistantproviderid() !==
                    apm.getAssistantprovideragentkit()?.getId() ? (
                      <IBlueBorderButton
                        disabled={
                          assistantProviderAction.assistant?.getAssistantproviderid() ===
                          apm.getAssistantprovideragentkit()?.getId()
                        }
                        className={cn(
                          'shrink-0 !bg-blue-600/10 hover:!bg-blue-600 border-none h-8 ring-[0.5] ring-blue-600',
                        )}
                        onClick={() => {
                          deployRevision(
                            'AGENTKIT',
                            apm.getAssistantprovideragentkit()?.getId()!,
                          );
                        }}
                      >
                        <CircleSlash
                          className="w-3.5 h-3.5 mr-2"
                          strokeWidth={1.5}
                        />
                        <span>Deploy revision</span>
                      </IBlueBorderButton>
                    ) : (
                      <RevisionIndicator
                        status={
                          assistantProviderAction.assistant?.getAssistantproviderid() ===
                          apm.getAssistantprovideragentkit()?.getId()
                            ? 'DEPLOYED'
                            : 'NOT_DEPLOYED'
                        }
                      />
                    )}
                  </TableCell>

                  <NameCell>
                    {
                      apm
                        .getAssistantprovideragentkit()
                        ?.getCreateduser()
                        ?.getName()!
                    }
                  </NameCell>
                  <DateCell
                    date={apm.getAssistantprovideragentkit()?.getCreateddate()}
                  ></DateCell>
                </TableRow>
              );
            case GetAllAssistantProviderResponse.AssistantProvider
              .AssistantproviderCase.ASSISTANTPROVIDERWEBSOCKET:
              return (
                <TableRow key={idx}>
                  <CopyCell>
                    {`vrsn_${apm.getAssistantproviderwebsocket()?.getId()}`}
                  </CopyCell>
                  <TableCell>
                    <AssistantProviderIndicator provider="websocket" />
                  </TableCell>
                  <TableCell>
                    {apm.getAssistantproviderwebsocket()?.getDescription()
                      ? apm.getAssistantproviderwebsocket()?.getDescription()
                      : 'Initial assistant version'}
                  </TableCell>
                  <TableCell>
                    {assistantProviderAction.assistant?.getAssistantproviderid() !==
                    apm.getAssistantproviderwebsocket()?.getId() ? (
                      <IBlueBorderButton
                        disabled={
                          assistantProviderAction.assistant?.getAssistantproviderid() ===
                          apm.getAssistantproviderwebsocket()?.getId()
                        }
                        className={cn(
                          'shrink-0 !bg-blue-600/10 hover:!bg-blue-600 border-none h-8',
                        )}
                        onClick={() => {
                          deployRevision(
                            'WEBSOCKET',
                            apm.getAssistantproviderwebsocket()?.getId()!,
                          );
                        }}
                      >
                        <CircleSlash
                          className="w-3.5 h-3.5 mr-2"
                          strokeWidth={1.5}
                        />
                        <span>Deploy revision</span>
                      </IBlueBorderButton>
                    ) : (
                      <RevisionIndicator
                        status={
                          assistantProviderAction.assistant?.getAssistantproviderid() ===
                          apm.getAssistantproviderwebsocket()?.getId()
                            ? 'DEPLOYED'
                            : 'NOT_DEPLOYED'
                        }
                      />
                    )}
                  </TableCell>
                  <NameCell>
                    {apm.getAssistantproviderwebsocket()?.getCreateduser() &&
                      apm
                        .getAssistantproviderwebsocket()
                        ?.getCreateduser()
                        ?.getName()!}
                  </NameCell>
                  <DateCell
                    date={apm.getAssistantproviderwebsocket()?.getCreateddate()}
                  ></DateCell>
                </TableRow>
              );
            default:
              return null;
          }
        })}
      </ScrollableResizableTable>
    </TableSection>
  );
}
