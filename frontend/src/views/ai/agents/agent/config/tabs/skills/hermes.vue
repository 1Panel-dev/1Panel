<template>
    <div>
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
                    <el-button :loading="loading" @click="loadInstalledSkills">
                        <el-icon><Refresh /></el-icon>
                    </el-button>
                </template>
                <template v-else>
                    <el-select v-model="marketSource" class="p-w-200" @change="handleMarketSourceChange">
                        <el-option
                            v-for="source in marketSourceOptions"
                            :key="source.value"
                            :value="source.value"
                            :label="source.label"
                        />
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
                <div v-if="groupedInstalledSkills.length" class="group-list">
                    <section v-for="group in groupedInstalledSkills" :key="group.key" class="group-section">
                        <div class="group-header">
                            <span class="group-title">{{ group.label }}</span>
                            <span class="group-count">{{ group.items.length }}</span>
                        </div>
                        <div class="skills-grid">
                            <el-card v-for="skill in group.items" :key="skill.name" class="skill-card">
                                <div class="skill-head">
                                    <div class="skill-name">{{ skill.name }}</div>
                                    <el-button
                                        v-permission
                                        v-if="skill.uninstallable"
                                        link
                                        type="danger"
                                        :loading="uninstallingSkill === skill.name"
                                        @click="uninstallSkill(skill)"
                                    >
                                        {{ t('commons.button.uninstall') }}
                                    </el-button>
                                </div>
                                <div v-if="skill.description || getInstalledSkillSummary(skill)" class="skill-desc">
                                    {{ skill.description || getInstalledSkillSummary(skill) }}
                                </div>
                                <div class="skill-tags">
                                    <el-tag v-if="skill.category" size="small" type="primary" effect="plain">
                                        {{ skill.category }}
                                    </el-tag>
                                    <el-tag
                                        v-for="tag in getInstalledSkillTags(skill)"
                                        :key="`${skill.name}-${tag}`"
                                        size="small"
                                        effect="plain"
                                    >
                                        {{ tag }}
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
                    <el-card v-for="skill in marketResults" :key="skill.identifier || skill.slug" class="skill-card">
                        <div class="skill-head">
                            <div>
                                <div class="skill-name">{{ skill.name || skill.slug }}</div>
                                <div class="skill-slug">{{ skill.identifier || skill.slug }}</div>
                            </div>
                            <el-button
                                v-permission
                                type="primary"
                                link
                                :loading="installingSkill === (skill.identifier || skill.slug)"
                                @click="installSkill(skill)"
                            >
                                {{ t('commons.button.install') }}
                            </el-button>
                        </div>
                        <div v-if="skill.description || skill.summary" class="skill-desc">
                            {{ skill.description || skill.summary }}
                        </div>
                        <div class="skill-tags">
                            <el-tag size="small" type="primary" effect="plain">
                                {{ skill.source || marketSource }}
                            </el-tag>
                            <el-tag v-if="skill.trust" size="small" effect="plain">
                                {{ skill.trust }}
                            </el-tag>
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
import { computed, ref, watch } from 'vue';
import { Refresh, Search } from '@element-plus/icons-vue';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { installAgentSkill, listAgentSkills, searchAgentSkills, uninstallAgentSkill } from '@/api/modules/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { listEnterprisePublishedSkillHub } from '@/extensions/xpack';
import { MsgSuccess } from '@/utils/message';
import { newUUID } from '@/utils/id';
import TaskLog from '@/components/log/task/index.vue';

type SkillViewMode = 'installed' | 'market';
type HermesSkillMarketSource = 'official' | 'skills-sh' | 'local-hub';

const { t } = useI18n();
const { isEE, isOffline } = useGlobalStore();
const loading = ref(false);
const searching = ref(false);
const mode = ref<SkillViewMode>('market');
const installedKeyword = ref('');
const marketKeyword = ref('');
const getDefaultMarketSource = (): HermesSkillMarketSource => {
    if (isEE.value && isOffline.value) {
        return 'local-hub';
    }
    return 'official';
};
const marketSource = ref<HermesSkillMarketSource>(getDefaultMarketSource());
const marketSearched = ref(false);
const agentId = ref(0);
const installedSkills = ref<AI.AgentSkillItem[]>([]);
const marketResults = ref<AI.AgentSkillSearchItem[]>([]);
const installingSkill = ref('');
const uninstallingSkill = ref('');
const taskLogRef = ref<InstanceType<typeof TaskLog>>();
const marketSourceOptions = computed(() => {
    if (isEE.value && isOffline.value) {
        return [
            { value: 'local-hub' as HermesSkillMarketSource, label: t('aiTools.agents.skillsMarketSourceLocalHub') },
        ];
    }
    const options = [
        { value: 'official' as HermesSkillMarketSource, label: t('aiTools.agents.skillsMarketSourceOfficial') },
        { value: 'skills-sh' as HermesSkillMarketSource, label: 'skills.sh' },
    ];
    if (isEE.value) {
        options.push({ value: 'local-hub', label: t('aiTools.agents.skillsMarketSourceLocalHub') });
    }
    return options;
});

function handleMarketSourceChange() {
    marketResults.value = [];
    marketSearched.value = false;
}

watch([isEE, isOffline], () => {
    const allowed = marketSourceOptions.value.map((option) => option.value);
    if (allowed.includes(marketSource.value)) {
        return;
    }
    marketSource.value = getDefaultMarketSource();
    handleMarketSourceChange();
});

const filteredInstalledSkills = computed(() => {
    const keyword = installedKeyword.value.trim().toLowerCase();
    if (!keyword) {
        return installedSkills.value;
    }
    return installedSkills.value.filter((item) =>
        [item.name, item.description, item.category, item.source, item.trust].join(' ').toLowerCase().includes(keyword),
    );
});

const groupedInstalledSkills = computed(() => {
    const installedExtensions: AI.AgentSkillItem[] = [];
    const builtInSkills: AI.AgentSkillItem[] = [];
    for (const skill of filteredInstalledSkills.value) {
        if (skill.uninstallable) {
            installedExtensions.push(skill);
            continue;
        }
        builtInSkills.push(skill);
    }
    return [
        {
            key: 'installed-extensions',
            label: t('app.installed'),
            items: installedExtensions,
        },
        {
            key: 'built-in',
            label: t('aiTools.agents.skillsGroupBuiltIn'),
            items: builtInSkills,
        },
    ].filter((group) => group.items.length > 0);
});

const getInstalledSkillSummary = (skill: AI.AgentSkillItem) => {
    return [skill.category, skill.source, skill.trust].filter(Boolean).join(' / ');
};

const getInstalledSkillTags = (skill: AI.AgentSkillItem) => {
    const tags = skill.tags?.slice(0, 3) || [];
    if (tags.length > 0) {
        return tags;
    }
    if (skill.source && skill.source === skill.trust) {
        return [skill.source];
    }
    return [skill.source, skill.trust].filter(Boolean);
};

const loadInstalledSkills = async () => {
    if (!agentId.value) {
        return;
    }
    loading.value = true;
    try {
        const res = await listAgentSkills({ agentId: agentId.value });
        installedSkills.value = res.data || [];
    } finally {
        loading.value = false;
    }
};

const searchMarketSkills = async () => {
    if (!agentId.value) {
        return;
    }
    if (marketSource.value !== 'local-hub' && !marketKeyword.value.trim()) {
        return;
    }
    searching.value = true;
    try {
        if (marketSource.value === 'local-hub') {
            const res = await listEnterprisePublishedSkillHub({
                agentType: 'hermes-agent',
                keyword: marketKeyword.value.trim(),
            });
            marketResults.value = (res?.data || []).map((item) => ({
                slug: String(item.id),
                identifier: String(item.id),
                name: item.name,
                description: item.description,
                summary: item.description,
                version: item.version,
                source: 'local-hub',
                trust: item.riskLevel,
                score: '',
            }));
            marketSearched.value = true;
            return;
        }
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

const handleModeChange = async (value: SkillViewMode) => {
    if (value !== 'installed') {
        return;
    }
    await loadInstalledSkills();
};

const load = async (id: number) => {
    agentId.value = id;
    mode.value = 'market';
    installedKeyword.value = '';
    marketKeyword.value = '';
    marketSource.value = getDefaultMarketSource();
    marketSearched.value = false;
    installedSkills.value = [];
    marketResults.value = [];
};

const installSkill = async (skill: AI.AgentSkillSearchItem) => {
    if (!agentId.value) {
        return;
    }
    const slug = skill.identifier || skill.slug;
    const taskID = newUUID();
    installingSkill.value = slug;
    try {
        await installAgentSkill({
            agentId: agentId.value,
            source: marketSource.value,
            slug,
            taskID,
        });
        taskLogRef.value?.openWithTaskID(taskID);
    } finally {
        installingSkill.value = '';
    }
};

const uninstallSkill = async (skill: AI.AgentSkillItem) => {
    if (!agentId.value) {
        return;
    }
    uninstallingSkill.value = skill.name;
    try {
        await uninstallAgentSkill({
            agentId: agentId.value,
            name: skill.name,
        });
        MsgSuccess(t('commons.msg.uninstallSuccess'));
        await loadInstalledSkills();
    } finally {
        uninstallingSkill.value = '';
    }
};

const handleTaskClose = async () => {
    await loadInstalledSkills();
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
    color: var(--el-text-color-secondary);
    line-height: 1.6;
    display: -webkit-box;
    overflow: hidden;
    text-overflow: ellipsis;
    line-clamp: 2;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
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
</style>
