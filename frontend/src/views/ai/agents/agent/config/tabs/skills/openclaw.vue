<template>
    <VersionSupport v-if="!supported" :min-version="openclawMinSupportedVersion" />
    <div v-else>
        <el-radio-group v-model="mode" class="view-switch" @change="handleModeChange">
            <el-radio-button label="market">{{ t('aiTools.agents.skillsMarket') }}</el-radio-button>
            <el-radio-button label="installed">{{ t('app.installed') }}</el-radio-button>
        </el-radio-group>

        <div v-loading="loading" class="skills-content">
            <div class="toolbar">
                <template v-if="mode === 'installed'">
                    <el-input
                        v-model="installedKeyword"
                        :placeholder="t('aiTools.agents.skillsSearchPlaceholder')"
                        clearable
                        class="search-input"
                    >
                        <template #prefix>
                            <el-icon><Search /></el-icon>
                        </template>
                    </el-input>
                    <el-button :loading="loading" @click="loadSkills">
                        <el-icon><Refresh /></el-icon>
                    </el-button>
                </template>
                <template v-else>
                    <el-select v-model="marketSource" class="p-w-200" @change="handleMarketSourceChange">
                        <el-option
                            :value="'clawhub-global'"
                            :label="t('aiTools.agents.skillsMarketSourceClawhubGlobal')"
                        />
                        <el-option :value="'clawhub-cn'" :label="t('aiTools.agents.skillsMarketSourceClawhubChina')" />
                        <el-option :value="'skillhub'" :label="t('aiTools.agents.skillsMarketSourceSkillhub')" />
                    </el-select>
                    <el-input
                        v-model="marketKeyword"
                        :placeholder="t('aiTools.agents.skillsSearchPlaceholder')"
                        clearable
                        class="search-input"
                        @keyup.enter="searchMarketSkills"
                    >
                        <template #prefix>
                            <el-icon><Search /></el-icon>
                        </template>
                    </el-input>
                    <el-button :loading="searching" @click="searchMarketSkills">
                        {{ t('commons.button.search') }}
                    </el-button>
                </template>
            </div>

            <template v-if="mode === 'installed'">
                <div v-if="groupedSkills.length" class="group-list">
                    <section v-for="group in groupedSkills" :key="group.key" class="group-section">
                        <div class="group-header">
                            <span class="group-title">{{ group.label }}</span>
                            <span class="group-count">{{ group.items.length }}</span>
                        </div>
                        <div class="skills-grid">
                            <el-card v-for="skill in group.items" :key="skill.name" class="skill-card">
                                <div class="skill-head">
                                    <div class="skill-name">{{ skill.name }}</div>
                                    <el-switch
                                        :model-value="!skill.disabled"
                                        :loading="updatingSkill === skill.name"
                                        @change="(value) => toggleSkill(skill, Boolean(value))"
                                    />
                                </div>
                                <el-tooltip
                                    placement="bottom-start"
                                    :show-after="200"
                                    popper-class="skill-desc-tooltip"
                                >
                                    <template #content>
                                        <div class="skill-desc-tooltip-content">{{ skill.description }}</div>
                                    </template>
                                    <div class="skill-desc">
                                        {{ skill.description }}
                                    </div>
                                </el-tooltip>
                                <div class="skill-tags">
                                    <el-tag size="small" type="primary" effect="plain">
                                        {{ group.tagLabel }}
                                    </el-tag>
                                </div>
                            </el-card>
                        </div>
                    </section>
                </div>
                <el-empty v-else :description="t('aiTools.agents.skillsEmpty')" />
            </template>

            <template v-else>
                <div v-if="marketResults.length" class="skills-grid">
                    <el-card v-for="skill in marketResults" :key="`${marketSource}-${skill.slug}`" class="skill-card">
                        <div class="skill-head">
                            <div>
                                <div class="skill-name">{{ skill.name || skill.slug }}</div>
                                <div class="skill-slug">{{ skill.slug }}</div>
                            </div>
                            <el-button
                                type="primary"
                                link
                                :loading="installingSkill === skill.slug"
                                @click="installSkill(skill)"
                            >
                                {{ t('commons.button.install') }}
                            </el-button>
                        </div>
                        <el-tooltip
                            v-if="skill.description || skill.summary"
                            placement="bottom-start"
                            :show-after="200"
                            popper-class="skill-desc-tooltip"
                        >
                            <template #content>
                                <div class="skill-desc-tooltip-content">{{ skill.description || skill.summary }}</div>
                            </template>
                            <div class="skill-desc">
                                {{ skill.description || skill.summary }}
                            </div>
                        </el-tooltip>
                        <div class="skill-meta">
                            <span v-if="skill.version">{{ `${t('app.version')}: ${skill.version}` }}</span>
                            <span v-if="skill.score">{{ `${t('aiTools.agents.skillsScore')}: ${skill.score}` }}</span>
                        </div>
                    </el-card>
                </div>
                <el-empty
                    v-else
                    :description="
                        marketSearched ? t('aiTools.agents.skillsMarketEmpty') : t('aiTools.agents.skillsMarketHint')
                    "
                />
            </template>
        </div>

        <TaskLog ref="taskLogRef" @close="handleTaskClose" />
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { Refresh, Search } from '@element-plus/icons-vue';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { installAgentSkill, listAgentSkills, searchAgentSkills, updateAgentSkill } from '@/api/modules/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgSuccess } from '@/utils/message';
import { newUUID } from '@/utils/id';
import { isOpenclawCurrentHTTPVersion } from '@/utils/agent';
import TaskLog from '@/components/log/task/index.vue';
import VersionSupport from '../components/version-support.vue';

type SkillGroupKey = 'builtIn' | 'external' | 'workspace' | 'extra' | 'other';
type SkillViewMode = 'installed' | 'market';
type SkillMarketSource = 'clawhub-global' | 'clawhub-cn' | 'skillhub';

const openclawMinSupportedVersion = '2026.3.23';
const props = defineProps<{
    appVersion: string;
}>();

const { t } = useI18n();
const { isIntl } = useGlobalStore();
const loading = ref(false);
const searching = ref(false);
const mode = ref<SkillViewMode>('market');
const installedKeyword = ref('');
const marketKeyword = ref('');
const getDefaultMarketSource = (): SkillMarketSource => (isIntl.value ? 'clawhub-global' : 'clawhub-cn');
const marketSource = ref<SkillMarketSource>(getDefaultMarketSource());
const marketSearched = ref(false);
const agentId = ref(0);
const skills = ref<AI.AgentSkillItem[]>([]);
const marketResults = ref<AI.AgentSkillSearchItem[]>([]);
const updatingSkill = ref('');
const installingSkill = ref('');
const taskLogRef = ref<InstanceType<typeof TaskLog>>();
const installedLoaded = ref(false);
const supported = computed(() => isOpenclawCurrentHTTPVersion(props.appVersion));

const groupTagLabels = computed<Record<SkillGroupKey, string>>(() => ({
    builtIn: t('aiTools.agents.skillsGroupBuiltIn'),
    external: t('aiTools.agents.skillsGroupExternal'),
    workspace: t('aiTools.agents.skillsGroupWorkspace'),
    extra: t('runtime.extension'),
    other: t('aiTools.agents.otherTab'),
}));

const formatSkillGroupLabel = (label: string) => {
    const suffix = t('aiTools.agents.skillsTab');
    return /[\u3040-\u30ff\u3400-\u9fff\uac00-\ud7af]/.test(suffix) ? `${label}${suffix}` : `${label} ${suffix}`;
};

const groupLabels = computed<Record<SkillGroupKey, string>>(() => ({
    builtIn: formatSkillGroupLabel(groupTagLabels.value.builtIn),
    external: formatSkillGroupLabel(groupTagLabels.value.external),
    workspace: formatSkillGroupLabel(groupTagLabels.value.workspace),
    extra: formatSkillGroupLabel(groupTagLabels.value.extra),
    other: formatSkillGroupLabel(groupTagLabels.value.other),
}));

const filteredSkills = computed(() => {
    const value = installedKeyword.value.trim().toLowerCase();
    if (!value) {
        return skills.value;
    }
    return skills.value.filter((item) =>
        [item.name, item.description, item.source].join(' ').toLowerCase().includes(value),
    );
});

const groupedSkills = computed(() => {
    const groups: Record<SkillGroupKey, AI.AgentSkillItem[]> = {
        builtIn: [],
        external: [],
        workspace: [],
        extra: [],
        other: [],
    };
    for (const skill of filteredSkills.value) {
        groups[resolveGroupKey(skill)].push(skill);
    }
    const order: SkillGroupKey[] = ['external', 'workspace', 'extra', 'other', 'builtIn'];
    return order
        .filter((key) => groups[key].length > 0)
        .map((key) => ({
            key,
            label: groupLabels.value[key],
            tagLabel: groupTagLabels.value[key],
            items: groups[key],
        }));
});

const resolveGroupKey = (skill: AI.AgentSkillItem): SkillGroupKey => {
    if (skill.bundled || skill.source === 'openclaw-bundled') {
        return 'builtIn';
    }
    if (skill.source === 'openclaw-managed') {
        return 'external';
    }
    if (skill.source === 'openclaw-workspace') {
        return 'workspace';
    }
    if (skill.source === 'openclaw-extra') {
        return 'extra';
    }
    return 'other';
};

const loadSkills = async () => {
    if (!supported.value || !agentId.value) {
        return;
    }
    loading.value = true;
    try {
        const res = await listAgentSkills({ agentId: agentId.value });
        skills.value = res.data || [];
        installedLoaded.value = true;
    } finally {
        loading.value = false;
    }
};

const searchMarketSkills = async () => {
    if (!supported.value || !agentId.value || !marketKeyword.value.trim()) {
        return;
    }
    searching.value = true;
    try {
        const res = await searchAgentSkills({
            agentId: agentId.value,
            source: marketSource.value,
            keyword: marketKeyword.value.trim(),
        });
        marketResults.value = res.data || [];
        marketSearched.value = true;
    } finally {
        searching.value = false;
    }
};

const handleMarketSourceChange = () => {
    marketResults.value = [];
    marketSearched.value = false;
};

const handleModeChange = async (value: SkillViewMode) => {
    if (value !== 'installed' || installedLoaded.value) {
        return;
    }
    await loadSkills();
};

const load = async (id: number) => {
    agentId.value = id;
    mode.value = 'market';
    marketSource.value = getDefaultMarketSource();
    marketResults.value = [];
    marketSearched.value = false;
    skills.value = [];
    installedLoaded.value = false;
};

const toggleSkill = async (skill: AI.AgentSkillItem, enabled: boolean) => {
    if (!supported.value || !agentId.value) {
        return;
    }
    updatingSkill.value = skill.name;
    try {
        await updateAgentSkill({
            agentId: agentId.value,
            name: skill.name,
            enabled,
        });
        MsgSuccess(t('aiTools.agents.saveSuccess'));
        await loadSkills();
    } finally {
        updatingSkill.value = '';
    }
};

const installSkill = async (skill: AI.AgentSkillSearchItem) => {
    if (!supported.value || !agentId.value) {
        return;
    }
    const taskID = newUUID();
    installingSkill.value = skill.slug;
    try {
        await installAgentSkill({
            agentId: agentId.value,
            source: skill.source as SkillMarketSource,
            slug: skill.slug,
            taskID,
        });
        taskLogRef.value?.openWithTaskID(taskID);
    } finally {
        installingSkill.value = '';
    }
};

const handleTaskClose = async () => {
    if (mode.value === 'installed') {
        await loadSkills();
        return;
    }
    installedLoaded.value = false;
};

defineExpose({
    load,
});
</script>

<style scoped lang="scss">
.view-switch {
    margin-bottom: 16px;
}

.skills-content {
    min-height: 200px;
}

.toolbar {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-bottom: 16px;
}

.search-input {
    flex: 1;
}

.group-list {
    display: flex;
    flex-direction: column;
    gap: 20px;
}

.group-section {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.group-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 14px;
    min-height: 42px;
    border: 1px solid var(--el-border-color);
    border-radius: 10px;
    background: var(--el-fill-color-light);
}

.group-title,
.group-count {
    color: var(--el-text-color-secondary);
}

.skills-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
    padding: 2px;
}

.skill-card {
    --el-card-border-color: var(--el-border-color-dark);
    border-radius: 12px;
}

.skill-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
}

.skill-name {
    font-weight: 600;
    font-size: 16px;
}

.skill-desc {
    margin-top: 12px;
    display: -webkit-box;
    overflow: hidden;
    text-overflow: ellipsis;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    color: var(--el-text-color-secondary);
    line-height: 1.6;
}

.skill-tags {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    margin-top: 12px;
}

.skill-slug {
    margin-top: 4px;
    color: var(--el-text-color-secondary);
    word-break: break-all;
}

.skill-meta {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    margin-top: 12px;
    color: var(--el-text-color-secondary);
}

:global(.skill-desc-tooltip) {
    max-width: 360px;
}

:global(.skill-desc-tooltip .skill-desc-tooltip-content) {
    white-space: normal;
    word-break: break-word;
}
</style>
