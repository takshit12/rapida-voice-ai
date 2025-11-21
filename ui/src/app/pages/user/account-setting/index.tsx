import { PageHeaderBlock } from '@/app/components/blocks/page-header-block';
import { PageTitleBlock } from '@/app/components/blocks/page-title-block';
import { SideTab } from '@/app/components/tab';
import { AccountSetting } from '@/app/pages/user/account-setting/account-setting';
import { NotificationSetting } from '@/app/pages/user/account-setting/notification-setting';
import { useGlobalNavigation } from '@/hooks/use-global-navigator';
import { cn } from '@/utils';
import { Bell, ChevronLeft, User2 } from 'lucide-react';

export const AccountSettingPage = () => {
  const { goToDashboard } = useGlobalNavigation();
  return (
    <>
      <PageHeaderBlock className="border-b">
        <div
          className="flex items-center gap-3 hover:text-red-600 hover:cursor-pointer"
          onClick={() => {
            goToDashboard();
          }}
        >
          <ChevronLeft className="w-5 h-5 mr-1" strokeWidth={1.5} />
          <PageTitleBlock className="font-medium text-[14.5px]">
            Back to Dashboard
          </PageTitleBlock>
        </div>
      </PageHeaderBlock>
      <div className="flex-1 flex h-full">
        <SideTab
          strict={false}
          active="Account"
          className={cn('w-64')}
          tabs={[
            {
              label: 'Account',
              labelIcon: <User2 className="w-4 h-4" strokeWidth={1.5} />,
              element: <AccountSetting />,
            },
            {
              label: 'Notification',
              labelIcon: <Bell className="w-4 h-4" strokeWidth={1.5} />,
              element: <NotificationSetting />,
            },
          ]}
        />
      </div>
    </>
  );
};
