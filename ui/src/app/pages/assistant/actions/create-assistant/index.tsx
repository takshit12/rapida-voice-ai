import { useState } from 'react';
import { Helmet } from '@/app/components/helmet';
import { useRapidaStore } from '@/hooks';
import { TabForm } from '@/app/components/form/tab-form';
import {
  IBlueBGArrowButton,
  IBlueButton,
  ICancelButton,
} from '@/app/components/form/button';
import {
  Assistant,
  ConnectionConfig,
  CreateAssistantProviderRequest,
  CreateAssistantRequest,
  GetAssistantResponse,
  Metadata,
} from '@rapidaai/react';
import { useConfirmDialog } from '@/app/pages/assistant/actions/hooks/use-confirmation';
import { useGlobalNavigation } from '@/hooks/use-global-navigator';
import { useCurrentCredential } from '@/hooks/use-credential';
import { ConfigPrompt } from '@/app/components/configuration/config-prompt';
import { randomMeaningfullName, randomString } from '@/utils';
import { FieldSet } from '@/app/components/form/fieldset';
import { FormLabel } from '@/app/components/form-label';
import { Input } from '@/app/components/form/input';
import { Textarea } from '@/app/components/form/textarea';
import { TagInput } from '@/app/components/form/tag-input';
import { AssistantTag } from '@/app/components/form/tag-input/assistant-tags';
import {
  GetDefaultTextProviderConfigIfInvalid,
  TextProvider,
  ValidateTextProviderDefaultOptions,
} from '@/app/components/providers/text';
import { BuildinToolConfig } from '@/app/components/tools';
import { Card, CardDescription, CardTitle } from '@/app/components/base/cards';
import { ExternalLink, Info, Plus, SquareFunction } from 'lucide-react';
import { ActionableEmptyMessage } from '@/app/components/container/message/actionable-empty-message';
import { ConfigureAssistantToolDialog } from '@/app/components/base/modal/assistant-configure-tool-modal';
import { YellowNoticeBlock } from '@/app/components/container/message/notice-block';
import { CardOptionMenu } from '@/app/components/menu';
import { CreateAssistant } from '@rapidaai/react';
import { CreateAssistantToolRequest } from '@rapidaai/react';
import { Struct } from 'google-protobuf/google/protobuf/struct_pb';
import { connectionConfig } from '@/configs';
import { ChatCompletePrompt } from '@/utils/prompt';
import toast from 'react-hot-toast/headless';
import { InputHelper } from '@/app/components/input-helper';
import { ConfigureAssistantNextDialog } from '@/app/components/base/modal/assistant-configure-next-modal';

/**
 *
 * @returns
 */
export function CreateAssistantPage() {
  /**
   * credentils and authentication parameters
   */
  const { authId, token, projectId } = useCurrentCredential();

  /**
   * global reloading
   */
  const { loading, showLoader, hideLoader } = useRapidaStore();

  /**
   * after creation of assistant maintaining stage
   */
  const [assistant, setAssistant] = useState<null | Assistant>(null);

  /**
   * navigation
   */
  const { goBack, goToAssistant } = useGlobalNavigation();

  /**
   *
   */
  const [createAssistantSuccess, setCreateAssistantSuccess] = useState(false);

  /**
   * multi step form
   */
  const [activeTab, setActiveTab] = useState<
    'choose-model' | 'tools' | 'define-assistant'
  >('choose-model');

  /**
   * Error message
   */
  const [errorMessage, setErrorMessage] = useState('');

  /**
   * Form fields
   */
  const [name, setName] = useState(randomMeaningfullName('assistant'));
  const [description, setDescription] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [tools, setTools] = useState<
    {
      name: string;
      description: string;
      fields: string;
      buildinToolConfig: BuildinToolConfig;
    }[]
  >([]);
  const [editingTool, setEditingTool] = useState<{
    name: string;
    description: string;
    fields: string;
    buildinToolConfig: BuildinToolConfig;
  } | null>(null);
  const [selectedModel, setSelectedModel] = useState<{
    provider: string;
    parameters: Metadata[];
  }>({
    provider: 'azure',
    parameters: GetDefaultTextProviderConfigIfInvalid('azure', []),
  });
  const [template, setTemplate] = useState<{
    prompt: { role: string; content: string }[];
    variables: { name: string; type: string; defaultvalue: string }[];
  }>({
    prompt: [{ role: 'system', content: '' }],
    variables: [],
  });
  const { showDialog, ConfirmDialogComponent } = useConfirmDialog({});
  const [configureToolOpen, setConfigureToolOpen] = useState(false);
  const onAddTag = (tag: string) => {
    setTags([...tags, tag]);
  };
  const onRemoveTag = (tag: string) => {
    setTags(tags.filter(t => t !== tag));
  };
  const onChangeProvider = (providerName: string) => {
    setSelectedModel({
      provider: providerName,
      parameters: GetDefaultTextProviderConfigIfInvalid(
        providerName,
        selectedModel.parameters,
      ),
    });
  };
  const onChangeParameter = (parameters: Metadata[]) => {
    setSelectedModel({ ...selectedModel, parameters });
  };

  /**
   *
   * @returns
   */
  const createAssistant = () => {
    showLoader('overlay');
    if (!name) {
      setErrorMessage('Please provide a valid name for assistant.');
      return false;
    }
    const assistantToolConfig = tools.map(t => {
      const req = new CreateAssistantToolRequest();
      req.setName(t.name);
      req.setDescription(t.description);
      req.setFields(Struct.fromJavaScript(JSON.parse(t.fields)));
      req.setExecutionmethod(t.buildinToolConfig.code);
      req.setExecutionoptionsList(t.buildinToolConfig.parameters);
      return req;
    });
    const assistantProvider = new CreateAssistantProviderRequest();
    const assistantModel =
      new CreateAssistantProviderRequest.CreateAssistantProviderModel();
    assistantModel.setTemplate(ChatCompletePrompt(template));
    assistantModel.setModelprovidername(selectedModel.provider);
    assistantModel.setAssistantmodeloptionsList(selectedModel.parameters);
    assistantProvider.setModel(assistantModel);
    const request = new CreateAssistantRequest();
    request.setAssistantprovider(assistantProvider);
    request.setAssistanttoolsList(assistantToolConfig);
    request.setName(name);
    request.setTagsList(tags);
    request.setDescription(description);
    CreateAssistant(
      connectionConfig,
      request,
      ConnectionConfig.WithDebugger({
        authorization: token,
        userId: authId,
        projectId: projectId,
      }),
    )
      .then((car: GetAssistantResponse) => {
        hideLoader();
        if (car?.getSuccess()) {
          let ast = car.getData();
          if (ast) {
            toast.success(
              'Assistant Created Successfully, Your AI assistant is ready to be deployed.',
            );
            setAssistant(ast);
            setCreateAssistantSuccess(true);
          }
        } else {
          const errorMessage =
            'Unable to create assistant. please try again later.';
          const error = car?.getError();
          if (error) {
            setErrorMessage(error.getHumanmessage());
            return;
          }
          setErrorMessage(errorMessage);
          return;
        }
      })
      .catch(er => {
        hideLoader();
        const errorMessage =
          'Unable to create assistant. please try again later.';
        setErrorMessage(errorMessage);
        return;
      });
  };

  /**
   * validate instruction
   * @returns
   */
  const validateInstruction = (): boolean => {
    setErrorMessage('');
    let err = ValidateTextProviderDefaultOptions(
      selectedModel.provider,
      selectedModel.parameters,
    );
    if (err) {
      setErrorMessage(err);
      return false;
    }

    // Add template prompt validation
    if (!template.prompt || template.prompt.length === 0) {
      setErrorMessage('Please provide a valid template prompt.');
      return false;
    }

    // Validate each prompt message in the template
    for (const message of template.prompt) {
      if (!message.role || !message.content || message.content.trim() === '') {
        setErrorMessage(
          'Each prompt message must have a valid role and non-empty content.',
        );
        return false;
      }
    }
    return true;
  };

  /**
   * validation of tools
   * @returns
   */
  const validateTool = (): boolean => {
    setErrorMessage('');
    if (tools.length === 0) {
      setErrorMessage('Please add atleast one tool for the assistant.');
      return false;
    }
    return true;
  };

  //
  return (
    <>
      <Helmet title="Create an assistant"></Helmet>
      <ConfirmDialogComponent />
      {assistant && (
        <ConfigureAssistantNextDialog
          assistant={assistant}
          modalOpen={createAssistantSuccess}
          setModalOpen={() => {
            setCreateAssistantSuccess(false);
            goToAssistant(assistant.getId());
          }}
        />
      )}

      <ConfigureAssistantToolDialog
        modalOpen={configureToolOpen}
        setModalOpen={v => {
          setEditingTool(null);
          setConfigureToolOpen(v);
        }}
        initialData={editingTool}
        onValidateConfig={updatedTool => {
          // Check for empty name
          if (!updatedTool.name.trim()) {
            return 'Please provide a valid tool name.';
          }

          // Check for duplicate name
          const isDuplicate = tools.some(
            tool =>
              tool.name !== editingTool?.name && tool.name === updatedTool.name,
          );

          if (isDuplicate) {
            return 'Please provide a unique tool name for tools.';
          }

          return null;
        }}
        onChange={updatedTool => {
          if (editingTool) {
            setTools(
              tools.map(tool =>
                tool.name === editingTool.name ? updatedTool : tool,
              ),
            );
          } else {
            setTools([...tools, updatedTool]);
          }
          setEditingTool(null);
          setConfigureToolOpen(false);
        }}
      />
      <TabForm
        formHeading="Complete all steps to create new assistant"
        activeTab={activeTab}
        onChangeActiveTab={() => {}}
        errorMessage={errorMessage}
        form={[
          {
            name: 'Configuration',
            description: 'Select the llm you want to use for your assistant.',
            code: 'choose-model',
            body: (
              <div className="">
                <YellowNoticeBlock className="flex items-center">
                  <Info className="shrink-0 w-4 h-4" />
                  <div className="ms-3 text-sm font-medium">
                    Rapida Assistant enables you to deploy intelligent
                    conversational agents across multiple channels.
                  </div>
                  <a
                    target="_blank"
                    href="https://doc.rapida.ai/assistants/overview"
                    className="h-7 flex items-center font-medium hover:underline ml-auto text-yellow-600"
                    rel="noreferrer"
                  >
                    Read documentation
                    <ExternalLink
                      className="shrink-0 w-4 h-4 ml-1.5"
                      strokeWidth={1.5}
                    />
                  </a>
                </YellowNoticeBlock>
                <div className="space-y-6 px-8 py-8 max-w-4xl">
                  <div className="space-y-6">
                    <TextProvider
                      onChangeParameter={onChangeParameter}
                      onChangeProvider={onChangeProvider}
                      parameters={selectedModel.parameters}
                      provider={selectedModel.provider}
                    />
                  </div>
                  <ConfigPrompt
                    instanceId={randomString(10)}
                    existingPrompt={template}
                    onChange={prompt => setTemplate(prompt)}
                  />
                </div>
              </div>
            ),
            actions: [
              <ICancelButton
                className="px-4 rounded-[2px]"
                onClick={() => showDialog(goBack)}
              >
                Cancel
              </ICancelButton>,
              <IBlueBGArrowButton
                type="button"
                isLoading={loading}
                className="px-4 rounded-[2px]"
                onClick={() => {
                  if (validateInstruction()) setActiveTab('tools');
                }}
              >
                Continue
              </IBlueBGArrowButton>,
            ],
          },

          {
            code: 'tools',
            name: 'Tools (optional)',
            description:
              'Let your assistant work with given differnt tools on behalf of you',
            body: (
              <div className="flex grow flex-col">
                <div className="flex items-center justify-between pl-4 bg-white dark:bg-gray-900 border-b">
                  Tool and MCPs
                  <div className="flex divide-x">
                    <IBlueButton
                      onClick={() => {
                        setConfigureToolOpen(true);
                      }}
                    >
                      Add another tool
                      <Plus className="w-4 h-4 ml-1.5" />
                    </IBlueButton>
                  </div>
                </div>
                <YellowNoticeBlock className="flex items-center">
                  <Info className="shrink-0 w-4 h-4" />
                  <div className="ms-3 text-sm font-medium">
                    Activate the tools you want your assistant to use, allowing
                    it to perform actions like fetching real-time data,
                    processing complex tasks, and more.
                  </div>
                </YellowNoticeBlock>
                {tools.length > 0 ? (
                  <div className="overflow-y-auto grid-cols-2 md:grid-cols-4 grid gap-2 px-4 py-2">
                    {tools.map((itm, idx) => {
                      const isMCP = itm.buildinToolConfig.code === 'mcp';

                      return (
                        <Card key={idx}>
                          <header className="flex justify-between items-start">
                            <div className="flex items-center gap-2">
                              <SquareFunction
                                className="w-7 h-7"
                                strokeWidth={1.5}
                              />
                              {isMCP && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300 font-medium">
                                  MCP
                                </span>
                              )}
                            </div>
                            <CardOptionMenu
                              options={[
                                {
                                  option: (
                                    <span className="text-red-600">
                                      Delete tool
                                    </span>
                                  ),
                                  onActionClick: () => {
                                    setTools(prevTools =>
                                      prevTools.filter(tool => tool !== itm),
                                    );
                                  },
                                },
                                {
                                  option: 'Edit tool',
                                  onActionClick: () => {
                                    setEditingTool(itm);
                                    setConfigureToolOpen(true);
                                  },
                                },
                              ]}
                              classNames="h-8 w-8 p-1 opacity-60"
                            />
                          </header>
                          <div className="flex-1 mt-3">
                            <CardTitle>{itm.name}</CardTitle>
                            <CardDescription>{itm.description}</CardDescription>
                            </CardDescription>
                          </div>
                        </Card>
                      );
                    })}
                  </div>
                ) : (
                  <div className="justify-self-center justify-center items-center mx-auto my-auto w-full">
                    <ActionableEmptyMessage
                      title="No Tools"
                      subtitle="There are no tools given added to the assistant"
                      action="Add a tool"
                      onActionClick={() => setConfigureToolOpen(true)}
                    />
                  </div>
                )}
              </div>
            ),
            actions: [
              <ICancelButton
                className="px-4 rounded-[2px]"
                onClick={() => showDialog(goBack)}
              >
                Cancel
              </ICancelButton>,
              <IBlueBGArrowButton
                type="button"
                isLoading={loading}
                onClick={() => {
                  if (tools.length === 0) {
                    setTools([]);
                    setErrorMessage('');
                    setActiveTab('define-assistant');
                    return;
                  }
                  if (validateTool()) setActiveTab('define-assistant');
                }}
                className="px-4 rounded-[2px]"
              >
                {tools.length === 0 ? 'Skip for now' : 'Continue'}
              </IBlueBGArrowButton>,
            ],
          },
          {
            code: 'define-assistant',
            name: 'Profile',
            description:
              'Provide the name, a brief description, and relevant tags for your assistant to help identify and categorize it.',
            actions: [
              <ICancelButton
                className="px-4 rounded-[2px]"
                onClick={() => showDialog(goBack)}
              >
                Cancel
              </ICancelButton>,
              <IBlueBGArrowButton
                isLoading={loading}
                type="button"
                onClick={createAssistant}
                className="px-4 rounded-[2px]"
              >
                Create assistant
              </IBlueBGArrowButton>,
            ],
            body: (
              <div className="space-y-6 px-8 py-8 max-w-4xl">
                <div className="h-fit pt-4 rounded-[2px] space-y-4">
                  <FieldSet>
                    <FormLabel>Name</FormLabel>
                    <Input
                      name="agent_name"
                      onChange={e => {
                        setName(e.target.value);
                      }}
                      value={name}
                      className="form-input"
                      placeholder="eg: your emotion detector"
                    ></Input>
                    <InputHelper>
                      Provide a name, that will appear in the assistant list and
                      help identify it.
                    </InputHelper>
                  </FieldSet>

                  <FieldSet>
                    <FormLabel>Description (Optional)</FormLabel>
                    <Textarea
                      row={5}
                      value={description}
                      placeholder={"What's the purpose of the assistant?"}
                      onChange={t => setDescription(t.target.value)}
                    />
                    <InputHelper>
                      Provide a description to explain what this assistant is
                      about.
                    </InputHelper>
                  </FieldSet>
                  <TagInput
                    tags={tags}
                    addTag={onAddTag}
                    removeTag={onRemoveTag}
                    allTags={AssistantTag}
                  />
                </div>
              </div>
            ),
          },
        ]}
      />
    </>
  );
}
