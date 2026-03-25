<template>
    <VersionSupport v-if="!supported" :min-version="openclawMinSupportedVersion" />
    <div v-else v-loading="loading">
        <div class="toolbar">
            <el-input
                v-model="keyword"
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
        </div>

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
                        <el-tooltip placement="top-start" :show-after="200" popper-class="skill-desc-tooltip">
                            <template #content>
                                <div class="skill-desc-tooltip-content">{{ skill.description }}</div>
                            </template>
                            <div class="skill-desc">
                                {{ skill.description }}
                            </div>
                        </el-tooltip>
                        <div class="skill-tags">
                            <el-tag size="small" effect="plain">
                                {{ group.tagLabel }}
                            </el-tag>
                        </div>
                    </el-card>
                </div>
            </section>
        </div>
        <el-empty v-else :description="t('aiTools.agents.skillsEmpty')" />
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { Refresh, Search } from '@element-plus/icons-vue';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { listAgentSkills, updateAgentSkill } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';
import { isOpenclawCurrentHTTPVersion } from '@/utils/agent';
import VersionSupport from './components/version-support.vue';

type SkillGroupKey = 'builtIn' | 'external' | 'workspace' | 'extra' | 'other';

const openclawMinSupportedVersion = '2026.3.23';
const props = defineProps<{
    appVersion: string;
}>();

const { t } = useI18n();
const loading = ref(false);
const keyword = ref('');
const agentId = ref(0);
const skills = ref<AI.AgentSkillItem[]>([]);
const updatingSkill = ref('');
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
    const value = keyword.value.trim().toLowerCase();
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
    const order: SkillGroupKey[] = ['external', 'extra', 'builtIn', 'workspace', 'other'];
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
    } finally {
        loading.value = false;
    }
};

const load = async (id: number) => {
    agentId.value = id;
    await loadSkills();
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

defineExpose({
    load,
});
</script>

<style scoped lang="scss">
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

:global(.skill-desc-tooltip) {
    max-width: 360px;
}

:global(.skill-desc-tooltip .skill-desc-tooltip-content) {
    white-space: normal;
    word-break: break-word;
}
</style>
