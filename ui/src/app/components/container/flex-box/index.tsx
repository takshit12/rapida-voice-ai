import { GeneralFooter } from '@/app/components/footer/general-footer';
import { Header } from '@/app/components/navigation/header';

interface FlexBoxProps extends React.HTMLAttributes<HTMLElement> {
  children: any;
  showFooter: boolean;
  isFloatingHeader: boolean;
}

export function FlexBox(props: FlexBoxProps) {
  const { children, showFooter, isFloatingHeader, ...attrs } = props;
  return (
    <main
      {...attrs}
      className="relative antialiased dark:text-gray-400 bg-gray-50 dark:bg-gray-950 h-screen flex flex-col flex-1"
    >
      <Header />
      <div className="flex flex-col flex-1 grow">{props.children}</div>
      <GeneralFooter></GeneralFooter>
    </main>
  );
}

FlexBox.defaultProps = {
  showFooter: false,
  isFloatingHeader: false,
};
