import { memo, useState } from 'react';
import { ChevronDownIcon } from '@/app/components/Icon/ChevronDown';
import { Disclosure } from '@/app/components/disclosure';
import { ChevronUpIcon } from '@/app/components/Icon/ChevronUp';
import { cn } from '@/utils';
import { EndpointIcon } from '@/app/components/Icon/Endpoint';
import { AssistantIcon } from '@/app/components/Icon/Assistant';
import { SidebarIconWrapper } from '@/app/components/navigation/sidebar/sidebar-icon-wrapper';
import { SidebarLabel } from '@/app/components/navigation/sidebar/sidebar-label';
import { SidebarSimpleListItem } from '@/app/components/navigation/sidebar/sidebar-simple-list-item';
import { useLocation } from 'react-router-dom';
import { Tooltip } from '@/app/components/tooltip';
import { BetaIcon } from '@/app/components/Icon/Beta';
import { BotMessageSquare, Box, Route } from 'lucide-react';

export const Deployment = memo(() => {
  const location = useLocation();
  const { pathname } = location;
  const [open, setOpen] = useState(false || pathname.includes('/deployment'));

  return (
    <li>
      <SidebarSimpleListItem
        className={cn('justify-between')}
        active={pathname.includes('/deployment')}
        onClick={() => {
          setOpen(!open);
        }}
        navigate="#"
      >
        <div className="flex items-center">
          <SidebarIconWrapper>
            <Box className={cn('w-5 h-5 opacity-75')} strokeWidth={1.5} />
          </SidebarIconWrapper>
          <SidebarLabel>Deployment</SidebarLabel>
        </div>
        <SidebarIconWrapper className="opacity-0 group-hover:opacity-100 transition-all duration-100">
          {open ? <ChevronUpIcon /> : <ChevronDownIcon />}
        </SidebarIconWrapper>
      </SidebarSimpleListItem>
      <Disclosure open={open}>
        <div className="ml-6 dark:border-gray-800 border-l hidden group-hover:block">
          <SidebarSimpleListItem
            className="mx-0 mr-2"
            active={pathname.includes('/deployment/endpoint')}
            navigate="/deployment/endpoint"
          >
            <SidebarIconWrapper>
              <Route className={cn('w-5 h-5 opacity-75')} strokeWidth={1.5} />
            </SidebarIconWrapper>
            <SidebarLabel>Endpoints</SidebarLabel>
          </SidebarSimpleListItem>

          <SidebarSimpleListItem
            className="mx-0 mr-2"
            active={pathname.includes('/deployment/assistant')}
            navigate="/deployment/assistant"
          >
            <SidebarIconWrapper>
              <BotMessageSquare
                className={cn('w-5 h-5 opacity-75')}
                strokeWidth={1.5}
              />
            </SidebarIconWrapper>
            <SidebarLabel>
              Assistants
              <Tooltip
                children={
                  <p className="text-xs">
                    We are working very hard <br />
                    to get you best experience.
                    <br />
                  </p>
                }
                icon={<BetaIcon />}
              ></Tooltip>
            </SidebarLabel>
          </SidebarSimpleListItem>
        </div>
      </Disclosure>
    </li>
  );
});
