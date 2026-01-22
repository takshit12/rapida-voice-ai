import { FC } from 'react';
import { Knowledge } from '@rapidaai/react';
import { InfoIcon } from 'lucide-react';
import { cn } from '@/utils';
import { Card } from '@/app/components/base/cards';
import { KnowledgeDropdown } from '@/app/components/dropdown/knowledge-dropdown';
import { FormLabel } from '@/app/components/form-label';
import CheckboxCard from '@/app/components/form/checkbox-card';
import { FieldSet } from '@/app/components/form/fieldset';
import { Input } from '@/app/components/form/input';
import { Slider } from '@/app/components/form/slider';
import { HybridSearchIcon } from '@/app/components/Icon/hybrid-search';
import { TextSearchIcon } from '@/app/components/Icon/text-search';
import { VectorSearchIcon } from '@/app/components/Icon/vector-search';
import { Tooltip } from '@/app/components/tooltip';
import { InputGroup } from '@/app/components/input-group';
import { RETRIEVE_METHOD } from '@/models/datasets';
import {
  ConfigureToolProps,
  ToolDefinitionForm,
  useParameterManager,
} from '../common';

// ============================================================================
// Constants
// ============================================================================

const SEARCH_TYPE_CONFIG = [
  {
    id: 'hybrid-search-type',
    value: RETRIEVE_METHOD.hybrid,
    icon: HybridSearchIcon,
    title: 'Hybrid Search',
    description:
      "Execute full-text search and vector searches simultaneously, re-rank to select the best match for the user's query.",
  },
  {
    id: 'vector-search-type',
    value: RETRIEVE_METHOD.semantic,
    icon: VectorSearchIcon,
    title: 'Semantic Search',
    description:
      'Generate query embeddings and search for the text chunk most similar to its vector representation.',
  },
  {
    id: 'text-search-type',
    value: RETRIEVE_METHOD.fullText,
    icon: TextSearchIcon,
    title: 'Full Text Search',
    description:
      'Index all terms in the document, allowing users to search any term and retrieve relevant text chunk containing those terms.',
  },
] as const;

// ============================================================================
// Main Component
// ============================================================================

export const ConfigureKnowledgeRetrieval: FC<ConfigureToolProps> = ({
  toolDefinition,
  onChangeToolDefinition,
  inputClass,
  onParameterChange,
  parameters,
}) => {
  const { getParamValue, updateParameter } = useParameterManager(
    parameters,
    onParameterChange,
  );

  return (
    <>
      <InputGroup title="Action Definition">
        <div className="flex flex-col gap-8 max-w-6xl">
          <KnowledgeDropdown
            className={cn('bg-light-background', inputClass)}
            currentKnowledge={getParamValue('tool.knowledge_id')}
            onChangeKnowledge={(knowledge: Knowledge) => {
              if (knowledge) {
                updateParameter('tool.knowledge_id', knowledge.getId());
              }
            }}
          />

          <FieldSet>
            <FormLabel>Retrieval setting</FormLabel>
            <div className="grid grid-cols-3 gap-3">
              {SEARCH_TYPE_CONFIG.map(config => (
                <SearchTypeCard
                  key={config.id}
                  {...config}
                  inputClass={inputClass}
                  isSelected={
                    getParamValue('tool.search_type') === config.value
                  }
                  onSelect={() =>
                    updateParameter('tool.search_type', config.value)
                  }
                />
              ))}
            </div>
          </FieldSet>

          <div className="grid grid-cols-2 w-full gap-4">
            <SliderField
              id="top_k"
              label="Top K"
              tooltip="Used to filter chunks that are most similar to user questions. The system will also dynamically adjust the value of Top K, according to max_tokens of the selected model."
              min={1}
              max={10}
              step={1}
              value={getParamValue('tool.top_k')}
              onChange={value => updateParameter('tool.top_k', value)}
              inputClass={inputClass}
            />

            <SliderField
              id="score_threshold"
              label="Score Threshold"
              tooltip="Used to filter chunks that are most similar to user questions. The system will also dynamically adjust the value of Top K, according to max_tokens of the selected model."
              min={0}
              max={1}
              step={0.1}
              inputStep={0.01}
              value={getParamValue('tool.score_threshold')}
              onChange={value => updateParameter('tool.score_threshold', value)}
              inputClass={inputClass}
            />
          </div>
        </div>
      </InputGroup>

      {toolDefinition && onChangeToolDefinition && (
        <ToolDefinitionForm
          toolDefinition={toolDefinition}
          onChangeToolDefinition={onChangeToolDefinition}
          inputClass={inputClass}
          documentationUrl="https://doc.rapida.ai/assistants/tools/add-knowledge-tool"
        />
      )}
    </>
  );
};

// ============================================================================
// Search Type Card
// ============================================================================

interface SearchTypeCardProps {
  id: string;
  value: string;
  icon: FC<{ className?: string }>;
  title: string;
  description: string;
  inputClass?: string;
  isSelected: boolean;
  onSelect: () => void;
}

const SearchTypeCard: FC<SearchTypeCardProps> = ({
  id,
  value,
  icon: Icon,
  title,
  description,
  inputClass,
  isSelected,
  onSelect,
}) => (
  <CheckboxCard
    type="radio"
    name="search-type"
    id={id}
    value={value}
    wrapperClassNames="h-auto"
    checked={isSelected}
    onChange={onSelect}
  >
    <Card
      className={cn(
        'p-3 flex flex-row space-x-3 bg-light-background h-auto',
        inputClass,
      )}
    >
      <div className="rounded-[2px] flex items-center justify-center bg-blue-200/30 dark:bg-blue-200/10 shrink-0 h-10 w-10">
        <Icon className="text-blue-600" />
      </div>
      <div className="flex flex-col">
        <span className="font-medium text-[14px]">{title}</span>
        <span className="text-sm opacity-80">{description}</span>
      </div>
    </Card>
  </CheckboxCard>
);

// ============================================================================
// Slider Field
// ============================================================================

interface SliderFieldProps {
  id: string;
  label: string;
  tooltip: string;
  min: number;
  max: number;
  step: number;
  inputStep?: number;
  value: string;
  onChange: (value: string) => void;
  inputClass?: string;
}

const SliderField: FC<SliderFieldProps> = ({
  id,
  label,
  tooltip,
  min,
  max,
  step,
  inputStep,
  value,
  onChange,
  inputClass,
}) => (
  <FieldSet className="flex justify-between">
    <FormLabel htmlFor={id}>
      {label}
      <Tooltip icon={<InfoIcon className="w-4 h-4 ml-1" />}>
        <p className={cn('font-normal text-sm p-1 w-64')}>{tooltip}</p>
      </Tooltip>
    </FormLabel>
    <div className="flex justify-between items-center space-x-2">
      <Slider
        min={min}
        max={max}
        step={step}
        value={value}
        onSlide={(val: number) => onChange(val.toString())}
      />
      <Input
        id={id}
        className={cn(
          'py-0 px-1 tabular-nums w-10 h-6 text-xs bg-light-background',
          inputClass,
        )}
        min={min}
        max={max}
        type="number"
        step={inputStep ?? step}
        value={value}
        onChange={e => onChange(e.target.value)}
      />
    </div>
  </FieldSet>
);
