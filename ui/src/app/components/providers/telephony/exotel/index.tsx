import { Metadata } from '@rapidaai/react';
import { FormLabel } from '@/app/components/form-label';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { InputHelper } from '@/app/components/input-helper';

export const ValidateExotelTelephonyOptions = (
  options: Metadata[],
): boolean => {
  const credentialID = options.find(
    opt => opt.getKey() === 'rapida.credential_id',
  );
  if (
    !credentialID ||
    !credentialID.getValue() ||
    credentialID.getValue().length === 0
  ) {
    return false;
  }
  // Validate language
  const phone = options.find(opt => opt.getKey() === 'phone');
  if (phone) {
    if (!phone.getValue() || phone.getValue().length === 0) {
      return false;
    }
  }
  return true;
};

export const ConfigureExotelTelephony: React.FC<{
  onParameterChange: (parameters: Metadata[]) => void;
  parameters: Metadata[] | null;
}> = ({ onParameterChange, parameters }) => {
  //
  const getParamValue = (key: string) =>
    parameters?.find(p => p.getKey() === key)?.getValue() ?? '';

  const updateParameter = (key: string, value: string) => {
    const updatedParams = [...(parameters || [])];
    const existingIndex = updatedParams.findIndex(p => p.getKey() === key);
    const newParam = new Metadata();
    newParam.setKey(key);
    newParam.setValue(value);
    if (existingIndex >= 0) {
      updatedParams[existingIndex] = newParam;
    } else {
      updatedParams.push(newParam);
    }
    onParameterChange(updatedParams);
  };

  return (
    <>
      <FieldSet className="col-span-1">
        <FormLabel>Phone</FormLabel>
        <Input
          className="bg-light-background"
          value={getParamValue('phone')}
          placeholder="Enter exotel phone number"
          onChange={v => {
            updateParameter('phone', v.target.value);
          }}
        />
        <InputHelper>
          Phone to recieve inbound or make outbound call.
        </InputHelper>
      </FieldSet>
      <FieldSet className="col-span-1">
        <FormLabel>App ID</FormLabel>
        <Input
          className="bg-light-background"
          value={getParamValue('app_id')}
          placeholder="Enter exotel applet app_id"
          onChange={v => {
            updateParameter('app_id', v.target.value);
          }}
        />
        <InputHelper>
          Exotel app_id is the identiﬁer of the flow (or applet) that you want
          to connect to once the From number picks up the call.
        </InputHelper>
      </FieldSet>
    </>
  );
};
