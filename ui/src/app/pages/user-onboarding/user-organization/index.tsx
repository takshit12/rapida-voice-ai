import React, { useCallback, useContext, useState } from 'react';
import { DescriptiveHeading } from '@/app/components/heading/descriptive-heading';
import { Helmet } from '@/app/components/helmet';
import { IBlueBGArrowButton } from '@/app/components/form/button';
import { Input } from '@/app/components/form/input';
import { Select } from '@/app/components/form/select';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { CreateOrganization } from '@rapidaai/react';
import { CreateOrganizationResponse } from '@rapidaai/react';
import { useCurrentCredential } from '@/hooks/use-credential';
import { useRapidaStore } from '@/hooks';
import { ErrorMessage } from '@/app/components/form/error-message';
import { ServiceError } from '@rapidaai/react';
import { AuthContext } from '@/context/auth-context';
import { FieldSet } from '@/app/components/form/fieldset';
import { FormLabel } from '@/app/components/form-label';
import { connectionConfig } from '@/configs';
export function CreateOrganizationPage() {
  /**
   * To naviagate to dashboard
   */
  const navigate = useNavigate();

  /**
   * setLoading context
   */
  const { loading, showLoader, hideLoader } = useRapidaStore();

  /**
   * Authenticaiton Context
   */
  const { authorize } = useContext(AuthContext);

  /**
   * credentials
   */
  const { user, authId, token } = useCurrentCredential();

  /**
   * handle the form
   */
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm();
  const [error, setError] = useState('');

  /**
   * orgnization options
   */
  const OrgOptions = [
    { name: 'Startup', value: 'startup' },
    { name: 'Late stage', value: 'late-stage' },
    { name: 'Enterprise', value: 'enterprise' },
  ];

  /**
   *
   * @param err
   * @param org
   * @returns
   */
  const afterCreateOrganization = useCallback(
    (err: ServiceError | null, org: CreateOrganizationResponse | null) => {
      if (err) {
        hideLoader();
        setError('Unable to process your request. please try again later.');
        return;
      }
      if (org?.getSuccess()) {
        authorize &&
          authorize(
            () => {
              hideLoader();
              return navigate('/onboarding/project');
            },
            () => {
              hideLoader();
              setError(
                'Please provide valid credentials to signin into account.',
              );
            },
          );
      } else {
        hideLoader();
        setError('Please provide valid credentials to signin into account.');
        return;
      }
    },
    [],
  );

  /**
   *
   * @param data
   */
  const onCreateOrganization = data => {
    showLoader('overlay');
    CreateOrganization(
      connectionConfig,
      data.organizationName,
      data.organizationSize,
      data.organizationIndustry,
      {
        authorization: token,
        'x-auth-id': authId,
      },
      afterCreateOrganization,
    );
  };
  return (
    <>
      <Helmet title="Onboarding: Create an organization"></Helmet>
      <DescriptiveHeading
        heading="Create your organization"
        subheading="This is where your team can work and collabrate on the projects."
      ></DescriptiveHeading>

      <form
        className="space-y-6 mt-6"
        onSubmit={handleSubmit(onCreateOrganization)}
      >
        <FieldSet>
          <FormLabel>Organization Name</FormLabel>
          <Input
            type="text"
            required
            defaultValue={`${user?.name}'s Organization`}
            placeholder="eg: Lexatic Inc"
            {...register('organizationName', {
              required: 'Please enter the organization name.',
            })}
          ></Input>
        </FieldSet>
        <FieldSet>
          <FormLabel>How large is your company?</FormLabel>
          <Select
            required
            placeholder="Select your orgnization size"
            {...register('organizationSize')}
            options={OrgOptions}
          ></Select>
        </FieldSet>
        <FieldSet>
          <FormLabel>Organization Industry</FormLabel>
          <Input
            required
            type="text"
            {...register('organizationIndustry', {
              required: 'Please provide industry for your organization.',
            })}
            placeholder="eg: Software, Engineering"
          ></Input>
        </FieldSet>
        <ErrorMessage
          message={
            (errors.organizationName?.message as string) ||
            (errors.organizationSize?.message as string) ||
            (errors.organizationIndustry?.message as string) ||
            error
          }
        />
        <IBlueBGArrowButton
          type="submit"
          className="w-full justify-between h-11"
          isLoading={loading}
        >
          Continue
        </IBlueBGArrowButton>
      </form>
    </>
  );
}
