import { VaultCredential } from '@rapidaai/react';
import { Dropdown } from '@/app/components/dropdown';
import { FormLabel } from '@/app/components/form-label';
import { IButton } from '@/app/components/form/button';
import { FieldSet } from '@/app/components/form/fieldset';
import { cn } from '@/utils';
import { Plus, RotateCcw } from 'lucide-react';
import { FC, useCallback, useEffect, useState } from 'react';
import { CreateProviderCredentialDialog } from '@/app/components/base/modal/create-provider-credential-modal';
import { useAllProviderCredentials } from '@/hooks/use-model';
import { useProviderContext } from '@/context/provider-context';
import { allProvider } from '@/providers';

interface CredentialDropdownProps {
  className?: string;
  provider?: string;
  currentCredential?: string;
  onChangeCredential: (credential: VaultCredential) => void;
}

export const CredentialDropdown: FC<CredentialDropdownProps> = props => {
  const [loading] = useState(false);
  const { providerCredentials } = useAllProviderCredentials();
  const ctx = useProviderContext();
  const [createProviderModalOpen, setCreateProviderModalOpen] = useState(false);
  const [currentProviderCredentials, setCurrentProviderCredentials] = useState<
    Array<VaultCredential>
  >([]);

  useEffect(() => {
    setCurrentProviderCredentials(
      providerCredentials.filter(y => y.getProvider() === props.provider),
    );
  }, [providerCredentials, props.provider]);
  const handleSearch = useCallback(
    (q: React.ChangeEvent<HTMLInputElement>) => {
      if (q.target.value && q.target.value.trim() !== '') {
        setCurrentProviderCredentials(
          providerCredentials.filter(
            y =>
              y.getProvider() === props.provider &&
              y.getName().includes(q.target.value.trim()),
          ),
        );
      } else {
        setCurrentProviderCredentials(
          providerCredentials.filter(y => y.getProvider() === props.provider),
        );
      }
    },
    [providerCredentials, props.provider],
  );

  return (
    <>
      <CreateProviderCredentialDialog
        modalOpen={createProviderModalOpen}
        setModalOpen={setCreateProviderModalOpen}
        currentProvider={props.provider}
      />
      <FieldSet>
        <FormLabel>Credential</FormLabel>
        <div
          className={cn(
            'outline-solid outline-transparent',
            'focus-within:outline-blue-600 focus:outline-blue-600',
            'border-b border-gray-300 dark:border-gray-700',
            'focus-within:border-transparent!',
            'transition-all duration-200 ease-in-out',
            'flex relative items-center',
            'bg-light-background dark:bg-gray-950 divide-x',
            'pt-px pl-px',
          )}
        >
          <div className="w-full relative">
            <Dropdown
              disable={loading}
              searchable
              className="
                bg-light-background dark:bg-gray-950 max-w-full focus-within:border-none! focus-within:outline-hidden! border-none! outline-hidden"
              currentValue={currentProviderCredentials.find(
                x => x.getId() === props.currentCredential,
              )}
              setValue={(c: VaultCredential) => {
                props.onChangeCredential(c);
              }}
              onSearching={handleSearch}
              allValue={currentProviderCredentials}
              placeholder="Select credential"
              option={(c: VaultCredential) => {
                return (
                  <div
                    className="relative overflow-hidden flex-1 flex flex-row space-x-3 py-1"
                    data-key={c.getId()}
                  >
                    <span className="inline-flex items-center gap-1.5 sm:gap-2 max-w-full text-sm font-medium">
                      <img
                        alt={
                          allProvider().find(x => x.code === c.getProvider())
                            ?.name
                        }
                        loading="lazy"
                        className="w-5 h-5 align-middle block shrink-0"
                        src={
                          allProvider().find(x => x.code === c.getProvider())
                            ?.image
                        }
                      />
                      <span className="truncate capitalize">
                        {
                          allProvider().find(x => x.code === c.getProvider())
                            ?.name
                        }
                      </span>
                      <span>/</span>
                      <span className="font-medium text-sm/6">
                        {c.getName()}
                      </span>
                    </span>
                  </div>
                );
              }}
              label={(c: VaultCredential) => {
                return (
                  <div className="relative overflow-hidden flex-1 flex flex-row space-x-3">
                    <div className="flex">
                      <span className="font-medium text-pretty text-sm/6">
                        {c.getName()}
                      </span>
                    </div>
                  </div>
                );
              }}
            />
          </div>
          <IButton
            className="bg-light-background dark:bg-gray-950 h-10"
            onClick={() => {
              ctx.reloadProviderCredentials();
            }}
          >
            <RotateCcw className={cn('w-4 h-4')} strokeWidth={1.5} />
          </IButton>
          <IButton
            className="bg-light-background dark:bg-gray-950 h-10"
            onClick={() => {
              setCreateProviderModalOpen(true);
            }}
          >
            <Plus className={cn('w-4 h-4')} strokeWidth={1.5} />
          </IButton>
        </div>
      </FieldSet>
    </>
  );
};
