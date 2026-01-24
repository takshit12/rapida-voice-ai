import { memo } from 'react';
import { cn } from '@/utils';
import { SidebarIconWrapper } from '@/app/components/navigation/sidebar/sidebar-icon-wrapper';
import { SidebarLabel } from '@/app/components/navigation/sidebar/sidebar-label';
import { SidebarSimpleListItem } from '@/app/components/navigation/sidebar/sidebar-simple-list-item';
import { useLocation } from 'react-router-dom';

import { Activity, Database, MessageSquareIcon, Webhook } from 'lucide-react';
import { ToolIcon } from '@/app/components/Icon/tool';

export const Observability = memo(() => {
  const location = useLocation();
  const { pathname } = location;

  return (
    <li>
      <SidebarSimpleListItem
        active={pathname.endsWith('/logs')}
        navigate="/logs"
      >
        <SidebarIconWrapper>
          <Activity className={cn('w-5 h-5 opacity-75')} strokeWidth={1.5} />
        </SidebarIconWrapper>
        <SidebarLabel>LLM logs</SidebarLabel>
      </SidebarSimpleListItem>

      <SidebarSimpleListItem
        active={pathname.includes('/logs/tool')}
        navigate="/logs/tool"
      >
        <SidebarIconWrapper>
          <ToolIcon className={cn('w-5 h-5 opacity-75')} strokeWidth={1.5} />
        </SidebarIconWrapper>
        <SidebarLabel>Tool logs</SidebarLabel>
      </SidebarSimpleListItem>
      <SidebarSimpleListItem
        active={pathname.includes('/logs/webhook')}
        navigate="/logs/webhook"
      >
        <SidebarIconWrapper>
          <Webhook className={cn('w-5 h-5 opacity-75')} strokeWidth={1.5} />
        </SidebarIconWrapper>
        <SidebarLabel>Webhook logs</SidebarLabel>
      </SidebarSimpleListItem>
      <SidebarSimpleListItem
        active={pathname.includes('/logs/knowledge')}
        navigate="/logs/knowledge"
      >
        <SidebarIconWrapper>
          <Database className={cn('w-5 h-5 opacity-75')} strokeWidth={1.5} />
        </SidebarIconWrapper>
        <SidebarLabel>Knowledge logs</SidebarLabel>
      </SidebarSimpleListItem>
      <SidebarSimpleListItem
        active={pathname.includes('/logs/conversation')}
        navigate="/logs/conversation"
      >
        <SidebarIconWrapper>
          <MessageSquareIcon
            className={cn('w-5 h-5 opacity-75')}
            strokeWidth={1.5}
          />
        </SidebarIconWrapper>
        <SidebarLabel>Conversation logs</SidebarLabel>
      </SidebarSimpleListItem>
    </li>
  );
});
