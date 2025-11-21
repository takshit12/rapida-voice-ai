import React, { FC, HTMLAttributes } from 'react';
import { cn } from '@/utils';

export const TableSection: FC<HTMLAttributes<HTMLDivElement>> = props => {
  return (
    <div className={cn('flex-1 flex flex-col', props.className)}>
      {props.children}
    </div>
  );
};
