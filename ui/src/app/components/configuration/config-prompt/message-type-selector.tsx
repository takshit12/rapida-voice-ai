import type { FC } from 'react';
import React from 'react';
import cn from 'classnames';
import { PromptRole } from '@/models/prompt';
import { Dropdown } from '@/app/components/dropdown';
type Props = {
  value?: PromptRole;
  onChange: (value: PromptRole) => void;
};

const allTypes = [PromptRole.system, PromptRole.user, PromptRole.assistant];
const MessageTypeSelector: FC<Props> = ({ value, onChange }) => {
  return (
    <Dropdown
      className={cn('min-w-[140px] !border-none !outline-none')}
      allValue={allTypes}
      currentValue={value}
      setValue={onChange}
      placeholder="Select a Role"
      label={cs => {
        return (
          <span className={cn('block truncate capitalize font-medium text-sm')}>
            {cs}
          </span>
        );
      }}
      option={(cs, selected) => {
        return (
          <span
            className={cn(
              'block truncate capitalize text-sm',
              selected ? 'opacity-100 font-medium' : 'opacity-80',
            )}
          >
            {cs}
          </span>
        );
      }}
    />
  );
};
export default React.memo(MessageTypeSelector);
