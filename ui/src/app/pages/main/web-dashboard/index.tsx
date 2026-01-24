import { MessageSquare, TestTube } from 'lucide-react';
import { ClickableCard } from '@/app/components/base/cards';
import { useCurrentCredential } from '@/hooks/use-credential';
import { KnowledgeIcon } from '@/app/components/Icon/knowledge';
import { EndpointIcon } from '@/app/components/Icon/Endpoint';
import { AssistantIcon } from '@/app/components/Icon/Assistant';
import { ModelIcon } from '@/app/components/Icon/Model';
import { IBlueBGArrowButton } from '@/app/components/form/button';
import { useGlobalNavigation } from '@/hooks/use-global-navigator';

export const HomePage = () => {
  const coreFeatures = [
    {
      icon: AssistantIcon,
      title: 'AI Assistants',
      description:
        'Deploy domain-specific AI agents with custom skills, workflows, and multi-step reasoning.',
      color: 'bg-green-500',
      route: '/deployment/assistant',
    },
    {
      icon: EndpointIcon,
      title: 'Governance & Endpoints',
      description:
        'Secure API endpoints with fine-grained governance, audit trails, and enterprise-grade access control.',
      color: 'bg-purple-500',
      route: '/deployment',
    },
    {
      icon: ModelIcon,
      title: 'Model Integration',
      description:
        'Bring your own model — support for OpenAI, Anthropic, and custom LLMs with fine-tuning capabilities.',
      color: 'bg-red-500',
      route: '/integration/models',
    },
    {
      icon: TestTube,
      title: 'Real-time Testing & Monitoring',
      description:
        'Instantly test AI agents and flows in a live sandbox to iterate faster and ship confidently.',
      color: 'bg-indigo-500',
      route: '/logs',
    },
    {
      icon: KnowledgeIcon,
      title: 'Knowledge Hub',
      description:
        'Unified repository for documents, training data, and AI knowledge management — the foundation of contextual intelligence.',
      color: 'bg-blue-500',
      route: '/knowledge',
    },
    {
      icon: MessageSquare,
      title: 'Conversational AI',
      description:
        'Context-aware, LLM-powered chat experiences that understand user intent and deliver accurate responses.',
      color: 'bg-yellow-500',
      route: '/deployment/assistant',
    },
  ];
  const { user } = useCurrentCredential();
  const { goToCreateAssistant } = useGlobalNavigation();
  return (
    <div className="flex-1 overflow-auto flex flex-col">
      {/* Core Platform Features */}

      <div className="border-b bg-white dark:bg-gray-900 p-4">
        <h1 className="text-xl font-semibold">
          Welcome, {user?.name?.split(/\s+/)[0] || user?.name}{' '}
        </h1>
      </div>
      <div className="bg-white dark:bg-gray-950">
        <div className="bg-blue-500/10 border-blue-500 border-b-[0.5px]  p-4 flex flex-col items-center text-center sm:flex-row sm:items-start sm:text-left">
          <div className="flex-1">
            <h2 className="mb-1 text-lg font-semibold">
              Design and Deploy Custom Voice Assistant
            </h2>
            <p className="mb-4 text-base">
              Build intelligent voice assistants that can handle calls, answer
              questions, and integrate with your existing systems. Deploy across
              websites, phone systems, or via SDK.
            </p>
            <div className="flex flex-col gap-3 sm:flex-row items-center space-x-2">
              <IBlueBGArrowButton
                onClick={() => {
                  goToCreateAssistant();
                }}
                className="text-base"
              >
                Create Assistant
              </IBlueBGArrowButton>
              <a
                rel="noreferrer"
                href="https://doc.rapida.ai/assistants/overview"
                target="_blank"
                className="text-base text-blue-600 font-medium hover:underline flex items-center space-x-1.5"
              >
                <span>View documentations</span>
              </a>
            </div>
          </div>
        </div>
      </div>

      <main className="px-6 py-6">
        <div className="grid grid-cols-1  lg:grid-cols-3 xl:grid-cols-5 gap-4">
          {coreFeatures.map((feature, index) => (
            <ClickableCard
              to={feature.route}
              key={index}
              className="transition-all duration-200 cursor-pointer group p-4 hover:shadow-lg border-b-2 hover:border-blue-600 h-full border col-span-1"
            >
              <div className="flex flex-col space-y-4">
                <div className="flex items-center justify-between">
                  <div
                    className={`w-10 h-10 ${feature.color} flex items-center justify-center group-hover:scale-110 transition-transform`}
                  >
                    <feature.icon
                      className="h-5 w-5 text-white"
                      strokeWidth={1.5}
                    />
                  </div>
                </div>
                <div>
                  <h3 className="text-lg font-semibold">{feature.title}</h3>
                  <p className="text-[0.95rem] text-gray-600 dark:text-gray-500 mt-2">
                    {feature.description}
                  </p>
                </div>
              </div>
            </ClickableCard>
          ))}
        </div>
      </main>
      <div className="border-y p-4 justify-between items-center bg-white dark:bg-gray-900 absolute bottom-0 w-full flex">
        <p>
          Reach out anytime — Get quick help from our team at:
          <a
            href="mailto:tech@rapida.ai"
            className="mx-2 text-blue-600 hover:underline underline-offset-2"
          >
            contact@rapida.ai
          </a>
        </p>
        <div className="flex">
          <p className="sm:px-3">© 2025 Rapida.ai. All rights reserved.</p>
          <a
            className="hover:text-gray-950 dark:hover:text-white sm:px-3"
            href="/static/privacy-policy"
          >
            Privacy Policy
          </a>
          <a
            className="hover:text-gray-950 dark:hover:text-white sm:px-3"
            href="/static/privacy-policy"
          >
            Terms and Conditions
          </a>
        </div>
      </div>
    </div>
  );
};
