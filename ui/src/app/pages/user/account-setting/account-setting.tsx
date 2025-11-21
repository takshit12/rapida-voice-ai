import { RedNoticeBlock } from '@/app/components/container/message/notice-block';
import { FormLabel } from '@/app/components/form-label';
import { IBlueBGArrowButton, IRedBGButton } from '@/app/components/form/button';
import { ErrorMessage } from '@/app/components/form/error-message';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { InputGroup } from '@/app/components/input-group';
import { InputHelper } from '@/app/components/input-helper';
import { connectionConfig } from '@/configs';
import { useRapidaStore } from '@/hooks';
import {
  ChangePassword,
  ChangePasswordRequest,
  ConnectionConfig,
} from '@rapidaai/react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import toast from 'react-hot-toast/headless';
import { useNavigate } from 'react-router-dom';
import { useCurrentCredential } from '@/hooks/use-credential';

export const AccountSetting = () => {
  /**
   * loggedin user
   */
  const { user, token, authId, projectId } = useCurrentCredential();
  /**
   * page error
   */
  const [error, setError] = useState('');

  /**
   * form handling
   */
  const { register, handleSubmit } = useForm();

  /**
   * common loader
   */
  const { loading, showLoader, hideLoader } = useRapidaStore();

  /**
   * To naviagate to dashboard
   */
  let navigate = useNavigate();

  /**
   * calling for authentication
   * @param data
   */
  const onChangePassword = data => {
    if (data.password !== data.re_password) {
      setError('The new passwords do not match. Please try again.');
    }
    showLoader();
    const request = new ChangePasswordRequest();
    request.setOldpassword(data.current_password);
    request.setPassword(data.password);
    ChangePassword(
      connectionConfig,
      request,
      ConnectionConfig.WithDebugger({
        authorization: token,
        userId: authId,
        projectId: projectId,
      }),
    )
      .then(rlp => {
        hideLoader();
        if (rlp?.getSuccess()) {
          toast.success(
            'The password has been successfully changed. You will be redirected to the sign-in page.',
          );
          return navigate('/auth/signin');
        } else {
          let errorMessage = rlp?.getError();
          if (errorMessage) setError(errorMessage.getHumanmessage());
          else {
            setError('Unable to process your request. please try again later.');
          }
          return;
        }
      })
      .catch(e => {
        setError('Unable to process your request. please try again later.');
        hideLoader();
      });
  };

  return (
    <div className="w-full flex flex-col flex-1">
      <div className="overflow-auto flex flex-col flex-1 pb-20">
        <InputGroup
          title="Account Information"
          className="bg-white dark:bg-gray-900"
        >
          <div className="space-y-6 max-w-md">
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
        <InputGroup title="Password" className="bg-white dark:bg-gray-900">
          <form
            className="space-y-6 max-w-lg"
            onSubmit={handleSubmit(onChangePassword)}
          >
            <FieldSet>
              <Input
                name="username"
                required
                type="hidden"
                value={user?.email}
              ></Input>
              <FormLabel>Current Password</FormLabel>
              <Input
                required
                type="password"
                autoComplete=""
                className="bg-light-background"
                {...register('current_password')}
                placeholder="*******"
              ></Input>
            </FieldSet>
            <FieldSet>
              <FormLabel>New Password</FormLabel>
              <Input
                required
                autoComplete="new-password"
                type="password"
                className="bg-light-background"
                {...register('re_password')}
                placeholder="*******"
              ></Input>
            </FieldSet>
            <FieldSet>
              <FormLabel>Confirm Password</FormLabel>
              <Input
                required
                type="password"
                autoComplete="new-password"
                className="bg-light-background"
                {...register('password')}
                placeholder="*******"
              ></Input>
            </FieldSet>
            <ErrorMessage message={error} />
            <IBlueBGArrowButton
              type="submit"
              isLoading={loading}
              className="px-4 rounded-[2px]"
            >
              Change Password
            </IBlueBGArrowButton>
          </form>
        </InputGroup>
        <InputGroup
          title="Account Deletion"
          initiallyExpanded={false}
          className="hidden"
        >
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
            >
              Yes, delete my account
            </IRedBGButton>
          </div>
        </InputGroup>
      </div>
    </div>
  );
};
