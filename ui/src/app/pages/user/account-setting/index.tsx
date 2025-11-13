import { PageHeaderBlock } from '@/app/components/blocks/page-header-block';
import { PageTitleBlock } from '@/app/components/blocks/page-title-block';
import { RedNoticeBlock } from '@/app/components/container/message/notice-block';
import { FormLabel } from '@/app/components/form-label';
import { IBlueBGButton, IRedBGButton } from '@/app/components/form/button';
import { InputCheckbox } from '@/app/components/form/checkbox';
import { ErrorMessage } from '@/app/components/form/error-message';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { InputGroup } from '@/app/components/input-group';
import { InputHelper } from '@/app/components/input-helper';
import { SideTab } from '@/app/components/tab';
import { useDeleteConfirmDialog } from '@/app/pages/assistant/actions/hooks/use-delete-confirmation';
import { useCurrentCredential } from '@/hooks/use-credential';
import { useGlobalNavigation } from '@/hooks/use-global-navigator';
import { cn } from '@/utils';
import { Bell, ChevronLeft, User2 } from 'lucide-react';
import { useState } from 'react';
export const AccountSettingPage = () => {
  const { goToDashboard } = useGlobalNavigation();
  return (
    <>
      <PageHeaderBlock className="border-b">
        <div
          className="flex items-center gap-3 hover:text-red-600 hover:cursor-pointer"
          onClick={() => {
            goToDashboard();
          }}
        >
          <ChevronLeft className="w-5 h-5 mr-1" strokeWidth={1.5} />
          <PageTitleBlock className="text-sm/6">
            Back to Dashboard
          </PageTitleBlock>
        </div>
      </PageHeaderBlock>
      <div className="flex-1 flex relative grow h-full overflow-hidden  bg-white dark:bg-gray-900">
        <SideTab
          strict={false}
          active="Account"
          className={cn('w-64')}
          tabs={[
            {
              label: 'Account',
              labelIcon: <User2 className="w-4 h-4" strokeWidth={1.5} />,
              element: <AccountSetting />,
            },
            {
              label: 'Notification',
              labelIcon: <Bell className="w-4 h-4" strokeWidth={1.5} />,
              element: <NotificationSetting />,
            },
          ]}
        />
      </div>
    </>
  );
};

const defaultNotifications = [
  {
    category: 'Assistant',
    items: [
      {
        id: 'assistant.created',
        label: 'Assistant Created',
        description: 'Triggered when a new assistant is created.',
        default: true,
      },
      {
        id: 'assistant.version.deployed',
        label: 'Assistant Deployed New Version',
        description:
          'Notifies when a new version of the assistant is deployed.',
        default: true,
      },
      {
        id: 'assistant.deleted',
        label: 'Assistant Deleted',
        description: 'Alert when an assistant is deleted.',
        default: false,
      },
      {
        id: 'assistant.deployment.created',
        label: 'Deployment Created',
        description:
          'Triggered when a new deployment is created for an assistant.',
        default: true,
      },
      {
        id: 'assistant.deployment.updated',
        label: 'Deployment Updated',
        description: 'Triggered when an assistant deployment is updated.',
        default: true,
      },
    ],
  },
];

const NotificationSetting = () => {
  const [errorMessage, setErrorMessage] = useState('');
  return (
    <>
      <form className="overflow-auto flex flex-col flex-1 pb-20">
        {defaultNotifications.map(notificationCategory => (
          <InputGroup
            title="Alert and Notifications"
            className="border-none"
            key={notificationCategory.category}
          >
            <hr className="border-t" />
            <div className="p-5 space-y-6">
              <div>
                <legend className="font-semibold ">
                  {notificationCategory.category} notifications
                </legend>
                <p className="mt-1 text-sm/6">
                  We'll always let you know about important changes, but you
                  pick what else you want to hear about.
                </p>
              </div>
              <div className="mt-6 space-y-6">
                {notificationCategory.items.map(item => (
                  <div className="flex gap-3" key={item.id}>
                    <div className="flex h-6 shrink-0 items-center">
                      <InputCheckbox
                        defaultChecked={item.default}
                        id={item.id}
                      />
                    </div>
                    <FieldSet className="text-sm/6">
                      <FormLabel htmlFor={item.id}>{item.label}</FormLabel>
                      <InputHelper id={`${item.id}-description`}>
                        {item.description}
                      </InputHelper>
                    </FieldSet>
                  </div>
                ))}
              </div>
              <IBlueBGButton
                type="button"
                //   isLoading={loading}
                //   onClick={onUpdateAssistantDetail}
                className="px-4 rounded-[2px]"
              >
                Update Notification
              </IBlueBGButton>
            </div>
          </InputGroup>
        ))}
      </form>
    </>
  );
};

const AccountSetting = () => {
  const { user } = useCurrentCredential();
  const [errorMessage, setErrorMessage] = useState('');
  const onUpdateAssistantDetail = () => {};

  // call it when you want to delete the assistant
  const Deletion = useDeleteConfirmDialog({
    onConfirm: () => {
      //   showLoader('block');
      //   const afterDeleteAssistant = (
      //     err: ServiceError | null,
      //     car: GetAssistantResponse | null,
      //   ) => {
      //     if (car?.getSuccess()) {
      //       toast.error('The assistant has been deleted successfully.');
      //       goToAssistantListing();
      //     } else {
      //       const error = car?.getError();
      //       if (error) {
      //         toast.error(error.getHumanmessage());
      //         return;
      //       }
      //       toast.error('Unable to delete assistant. please try again later.');
      //       return;
      //     }
      //   };
      //   DeleteAssistant(connectionConfig, assistantId, afterDeleteAssistant, {
      //     authorization: token,
      //     'x-auth-id': authId,
      //     'x-project-id': projectId,
      //   });
    },
    name: 'love',
  });

  return (
    <div className="w-full flex flex-col flex-1">
      <Deletion.ConfirmDeleteDialogComponent />
      <div className="overflow-auto flex flex-col flex-1 pb-20">
        <InputGroup title="Account Information" className="border-none">
          <hr className="border-t" />
          <div className="p-5 space-y-6 max-w-md">
            <FieldSet>
              <FormLabel>Name</FormLabel>
              <Input
                disabled
                className="bg-light-background"
                value={user?.name}
                placeholder="eg: John Deo"
              ></Input>
            </FieldSet>
            <FieldSet>
              <FormLabel>Email</FormLabel>
              <Input
                disabled
                className="bg-light-background"
                value={user?.email}
                placeholder="eg: john@rapida.ai"
              ></Input>
            </FieldSet>
          </div>
        </InputGroup>
        <InputGroup title="Password">
          <form className="p-5 space-y-6 max-w-lg">
            <FieldSet>
              <FormLabel>Current Password</FormLabel>
              <Input
                name="usecase"
                className="bg-light-background"
                onChange={e => {
                  // setName(e.target.value);
                }}
                //   value={name}
                placeholder="*******"
              ></Input>
            </FieldSet>
            <FieldSet>
              <FormLabel>New Password</FormLabel>
              <Input
                name="usecase"
                className="bg-light-background"
                onChange={e => {
                  // setName(e.target.value);
                }}
                //   value={name}
                placeholder="*******"
              ></Input>
            </FieldSet>
            <FieldSet>
              <FormLabel>Confirm Password</FormLabel>
              <Input
                name="usecase"
                className="bg-light-background"
                onChange={e => {
                  // setName(e.target.value);
                }}
                //   value={name}
                placeholder="*******"
              ></Input>
            </FieldSet>
            <ErrorMessage message={''} />
            <IBlueBGButton
              type="submit"
              // isLoading={loading}
              onClick={onUpdateAssistantDetail}
              className="px-4 rounded-[2px]"
            >
              Change Password
            </IBlueBGButton>
          </form>
        </InputGroup>
        <InputGroup title="Account Deletion" initiallyExpanded={false}>
          <RedNoticeBlock>
            Active connections will be terminated immediately, and the data will
            be permanently deleted after the rolling period.
          </RedNoticeBlock>
          <div className="flex flex-row items-center justify-between p-6">
            <FieldSet>
              <p className="font-semibold">Delete this account</p>
              <InputHelper className="-mt-1">
                No longer want to use our service? You can delete your account
                here. This action is not reversible. All information related to
                this account will be deleted permanently.
              </InputHelper>
            </FieldSet>
            <IRedBGButton
              className="rounded-[2px] font-medium text-sm/6"
              // isLoading={loading}
              onClick={Deletion.showDialog}
            >
              Yes, delete my account
            </IRedBGButton>
          </div>
        </InputGroup>
      </div>
    </div>
  );
};
