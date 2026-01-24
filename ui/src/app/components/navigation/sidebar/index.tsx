import { Observability } from '@/app/components/navigation/sidebar/observability';
import { Deployment } from '@/app/components/navigation/sidebar/deployment';
import { Dashboard } from '@/app/components/navigation/sidebar/dashboard';
import { Team } from '@/app/components/navigation/sidebar/team';
import { Project } from '@/app/components/navigation/sidebar/project';
import { Vault } from '@/app/components/navigation/sidebar/vault';
import { SidebarLabel } from '@/app/components/navigation/sidebar/sidebar-label';
import { Knowledge } from '@/app/components/navigation/sidebar/knowledge';
import { Aside } from '@/app/components/aside';
import { ExternalTool } from '@/app/components/navigation/sidebar/external-tools';
import { useWorkspace } from '@/workspace';
import { RapidaIcon } from '@/app/components/Icon/Rapida';
import { RapidaTextIcon } from '@/app/components/Icon/RapidaText';
import { ChevronsLeft } from 'lucide-react';
import { Tooltip } from '@/app/components/tooltip';
import { useSidebar } from '@/context/sidebar-context';
import { cn } from '../../../../utils/index';

/**
 *
 * @param props
 * @returns
 */
export function SidebarNavigation(props: {}) {
  const workspace = useWorkspace();
  const { locked, setLocked, open } = useSidebar();
  return (
    <Aside className="space-y-2 relative shrink-0">
      <div className="flex rounded-[2px] my-2 text-blue-600 items-center">
        <div className="pl-[0.8rem] py-2.5 shrink-0">
          {workspace.logo ? (
            <>
              <img
                src={workspace.logo.light}
                alt={workspace.title}
                className="h-8 block dark:hidden"
              />
              <img
                src={workspace.logo.dark}
                alt={workspace.title}
                className="h-8 hidden dark:block"
              />
            </>
          ) : (
            <div className="flex items-center shrink-0 space-x-1.5 ml-1 text-blue-600 dark:text-blue-500">
              <RapidaIcon className="h-8 w-8" />
              <RapidaTextIcon className="h-5" />
            </div>
          )}
        </div>
      </div>
      <div className="space-y-4">
        <ul className="space-y-1">
          <Dashboard />
          <Deployment />
          <Knowledge />
        </ul>
        <div className="space-y-3">
          <SidebarLabel
            className={cn(
              'uppercase truncate pl-3 text-xs opacity-0',
              open ? 'opacity-100' : 'opacity-0',
            )}
          >
            Observability
          </SidebarLabel>
          <ul className="space-y-1 mt-2">
            <Observability />
          </ul>
        </div>
        <div className="space-y-3">
          <SidebarLabel
            className={cn(
              'uppercase truncate pl-3 text-xs opacity-0',
              open ? 'opacity-100' : 'opacity-0',
            )}
          >
            Integrations
          </SidebarLabel>
          <ul className="space-y-1 mt-2">
            <ExternalTool />
            <Vault />
          </ul>
        </div>
        <div className="space-y-3">
          <SidebarLabel
            className={cn(
              'uppercase truncate pl-3 text-xs opacity-0',
              open ? 'opacity-100' : 'opacity-0',
            )}
          >
            Organizations
          </SidebarLabel>
          <ul className="space-y-1  mt-2">
            <Team />
            <Project />
          </ul>
        </div>
      </div>
      <div
        className="absolute bottom-0 right-0 w-10 h-10"
        onClick={() => setLocked(!locked)}
      >
        <Tooltip
          icon={
            <ChevronsLeft
              className={cn(
                'w-6 h-6 transition-all delay-200',
                !locked && 'rotate-180',
              )}
              strokeWidth={1.5}
            />
          }
        >
          <span className="text-gray-600 dark:text-gray-300">
            Lock sidebar open
          </span>
        </Tooltip>
      </div>
    </Aside>
  );
}
