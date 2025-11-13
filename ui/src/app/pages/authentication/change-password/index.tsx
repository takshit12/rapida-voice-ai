import React, { useCallback, useState } from 'react';
import { DescriptiveHeading } from '@/app/components/heading/descriptive-heading';
import { Helmet } from '@/app/components/helmet';
import { Input } from '@/app/components/form/input';
import { Label } from '@/app/components/form/label';
import { useNavigate, useParams } from 'react-router-dom';
import { CreatePassword } from '@rapidaai/react';
import { CreatePasswordResponse } from '@rapidaai/react';
import { useForm } from 'react-hook-form';
import { IBlueBGArrowButton } from '@/app/components/form/button';
import { ErrorMessage } from '@/app/components/form/error-message';
import { ServiceError } from '@rapidaai/react';
import { connectionConfig } from '@/configs';
import { useRapidaStore } from '@/hooks';
import { FieldSet } from '@/app/components/form/fieldset';

/**
 *
 * @returns
 */
export function ChangePasswordPage() {
  /**
   * Form handling
   */
  const { register, handleSubmit } = useForm();
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const { token } = useParams();
  const { loading, showLoader, hideLoader } = useRapidaStore();

  /**
   * after changing the password
   */
  const afterCreatePassword = useCallback(
    (err: ServiceError | null, cpr: CreatePasswordResponse | null) => {
      hideLoader();
      if (err) {
        setError('unable to process your request. please try again later.');
        return;
      }
      if (cpr?.getSuccess()) {
        return navigate('/auth/signin');
      } else {
        let errorMessage = cpr?.getError();
        if (errorMessage) setError(errorMessage.getHumanmessage());
        else
          setError('Unable to process your request. please try again later.');
        return;
      }
    },
    [],
  );
  /**
   *
   * @param data
   * @returns
   */
  const onCreatePassword = data => {
    if (!token) {
      setError(
        'The password token is expired, please request again for reset password token.',
      );
      return;
    }
    if (data.password !== data.confirmPassword) {
      setError('Passwords entered do not match, please check and try again.');
      return;
    }
    showLoader();
    CreatePassword(connectionConfig, token, data.password, afterCreatePassword);
  };

  return (
    <>
      <Helmet title="Forgot your password"></Helmet>
      <DescriptiveHeading
        heading="Change Password"
        subheading="You’ve requested to change your password. Please enter your new password below to secure your account. Once updated, you can use your new password to sign in."
      ></DescriptiveHeading>
      <form
        className="space-y-6 mt-6"
        onSubmit={handleSubmit(onCreatePassword)}
      >
        <FieldSet>
          <Label for="password" text="Password"></Label>
          <Input
            required
            {...register('password')}
            type="password"
            placeholder="********"
          ></Input>
        </FieldSet>
        <FieldSet>
          <Label for="password" text="Confirm Password"></Label>
          <Input
            required
            {...register('confirmPassword')}
            type="password"
            placeholder="********"
          ></Input>
        </FieldSet>
        <ErrorMessage message={error} />
        <IBlueBGArrowButton
          type="submit"
          className="w-full justify-between h-11"
          isLoading={loading}
        >
          Change Password
        </IBlueBGArrowButton>
      </form>
    </>
  );
}
