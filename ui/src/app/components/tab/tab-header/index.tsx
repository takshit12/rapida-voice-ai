import { cn } from '@/utils';
import { HTMLAttributes } from 'react';

export function TabHeader(props: HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn('border-b dark:border-gray-800 h-10', props.className)}>
      {props.children}
    </div>
  );
}
