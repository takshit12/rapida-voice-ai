import { FC, HTMLAttributes } from 'react';
import { Card, CardDescription, CardTitle } from '@/app/components/base/cards';
import { cn } from '@/utils';
import { CardOptionMenu } from '@/app/components/menu';
import { AssistantTool } from '@rapidaai/react';
import { BUILDIN_TOOLS } from '@/llm-tools';

interface ToolCardProps extends HTMLAttributes<HTMLDivElement> {
  tool: AssistantTool;
  options?: { option: any; onActionClick: () => void }[];
  iconClass?: string; // Fixed typo from 'iconClasss'
  titleClass?: string;
  isConnected?: boolean;
}

export const SelectToolCard: FC<ToolCardProps> = ({
  tool,
  options,
  className,
}) => {
  // Safely check if tool has protobuf methods
  const hasProtobufMethods = typeof tool.getExecutionmethod === 'function';
  const executionMethod = hasProtobufMethods ? tool.getExecutionmethod() : '';
  const isMCP = executionMethod === 'mcp';

  const toolName = hasProtobufMethods ? tool.getName?.() : (tool as any).name;
  const toolDescription = hasProtobufMethods
    ? tool.getDescription?.()
    : (tool as any).description;

  return (
    <Card className={cn(className)}>
      <header className="flex justify-between items-start">
        <div className="flex items-center gap-2">
          <img
            alt={BUILDIN_TOOLS.find(x => x.code === executionMethod)?.name}
            src={BUILDIN_TOOLS?.find(x => x.code === executionMethod)?.icon}
            className="w-7 h-7 mr-1.5 inline-block"
          />
          {isMCP && (
            <span className="text-[10px] px-1.5 py-0.5 rounded bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300 font-medium">
              MCP
            </span>
          )}
        </div>
        {options && (
          <CardOptionMenu
            options={options}
            classNames="h-8 w-8 p-1 opacity-60"
          />
        )}
      </header>
      <div className="flex-1 mt-3">
        <CardTitle>{toolName}</CardTitle>
        <CardDescription>{toolDescription}</CardDescription>
      </div>
    </Card>
  );
};
