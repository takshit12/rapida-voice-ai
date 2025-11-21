import { Dropdown } from '@/app/components/dropdown';
import { INTEGRATION_PROVIDER, RapidaProvider } from '@/providers';
import { FC } from 'react';

/**
 * all the props for dropdown
 */
interface ProviderDropdownProps {
  currentProvider?: RapidaProvider;
  setCurrentProvider: (v: RapidaProvider) => void;
}

/**
 *
 * @param props
 * @returns
 */
export const ProviderDropdown: FC<ProviderDropdownProps> = props => {
  const dropdownItem = (p: RapidaProvider) => {
    return (
      <span className="inline-flex items-center gap-1.5 sm:gap-2 max-w-full text-sm font-medium">
        <img
          alt={p.name}
          loading="lazy"
          className="w-5 h-5 align-middle block shrink-0"
          src={p.image}
        />
        <span className="truncate capitalize">{p.name}</span>
      </span>
    );
  };
  return (
    <Dropdown
      currentValue={props.currentProvider}
      setValue={props.setCurrentProvider}
      allValue={INTEGRATION_PROVIDER}
      className="bg-light-background dark:bg-gray-950"
      placeholder="Select the provider"
      label={dropdownItem}
      option={dropdownItem}
    />
  );
};
