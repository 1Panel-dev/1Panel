<template>
    <el-cascader
        :model-value="displayValue"
        :options="options"
        :props="{ multiple: true, emitPath: false }"
        :show-all-levels="true"
        clearable
        filterable
        class="w-full"
        :disabled="loading"
        @change="changeSelection"
    >
        <template #tag="{ data }">
            <el-tag
                v-for="tag in selectedTags(data).slice(0, 3)"
                :key="String(tag.node.value)"
                type="info"
                :closable="tag.closable"
                disable-transitions
                @close="removeSelection(String(tag.node.value))"
            >
                {{ tag.text }}
            </el-tag>
            <el-tooltip v-if="selectedTags(data).length > 3" placement="bottom">
                <template #content>
                    <div v-for="tag in selectedTags(data).slice(3)" :key="String(tag.node.value)">
                        {{ tag.text }}
                    </div>
                </template>
                <el-tag type="info" disable-transitions>+ {{ selectedTags(data).length - 3 }}</el-tag>
            </el-tooltip>
        </template>
        <template #default="{ node, data }">
            <span :class="{ 'website-group-label': !node.isLeaf }">{{ data.label }}</span>
        </template>
    </el-cascader>
    <span class="input-help">{{ $t('cronjob.websiteGroupDynamicHelper') }}</span>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { getWebsiteOptions } from '@/api/modules/website';
import { getAgentGroupList } from '@/api/modules/group';
import i18n from '@/lang';
import { Website } from '@/api/interface/website';

const props = defineProps<{ modelValue: string[] }>();
const emit = defineEmits<{ (event: 'update:modelValue', value: string[]): void }>();
const websites = ref<Website.WebsiteOption[]>([]);
const groups = ref<{ id: number; name: string }[]>([]);
const loading = ref(true);
const members = (id: number) =>
    websites.value.filter((site) => site.webSiteGroupId === id).map((site) => String(site.id));
const panelOrder = ref<string[]>([]);
const displayValue = computed(() => {
    const values = new Set(props.modelValue);
    for (const group of groups.value) {
        if (values.has(`group:${group.id}`)) members(group.id).forEach((id) => values.add(id));
    }
    return [...panelOrder.value.filter((value) => values.delete(value)), ...values];
});
// Expanded website values keep the panel checked, but only persisted selections become tags.
const tagValues = computed(() => {
    const selected = new Set(props.modelValue);
    for (const group of groups.value) {
        if (selected.has(`group:${group.id}`)) members(group.id).forEach((id) => selected.delete(id));
    }
    return selected;
});
const selectedTags = (tags: { node: { value: unknown }; text: string; closable: boolean }[]) =>
    tags.filter((tag) => tagValues.value.has(String(tag.node.value)));
const removeSelection = (value: string) => {
    const removed = new Set([value]);
    if (value.startsWith('group:')) members(Number(value.slice(6))).forEach((id) => removed.add(id));
    emit(
        'update:modelValue',
        props.modelValue.filter((item) => !removed.has(item)),
    );
};
type WebsiteOption = { value: string; label: string; children?: WebsiteOption[] };
const availableOptions = computed(() => {
    const all = i18n.global.t('commons.table.all');
    const result: WebsiteOption[] = [{ value: 'all', label: all }];
    const known = new Set(['all']);
    for (const group of groups.value) {
        const children = [{ value: `group:${group.id}`, label: all }];
        known.add(`group:${group.id}`);
        for (const site of websites.value.filter((site) => site.webSiteGroupId === group.id)) {
            children.push({ value: String(site.id), label: `${site.primaryDomain} (${site.alias})` });
            known.add(String(site.id));
        }
        result.push({ value: `menu:${group.id}`, label: group.name, children });
    }
    // Preserve sites without a matching group, and stale saved values, for explicit removal.
    for (const site of websites.value.filter((site) => !known.has(String(site.id)))) {
        result.push({ value: String(site.id), label: site.primaryDomain });
        known.add(String(site.id));
    }
    return { nodes: result, known };
});
// Only missing saved selections depend on modelValue; normal selection changes leave the tree intact.
const missingValues = computed<string[]>((previous) => {
    const values = props.modelValue.filter((value) => !availableOptions.value.known.has(value));
    return previous && previous.length === values.length && previous.every((value, index) => value === values[index])
        ? previous
        : values;
});
const options = computed<WebsiteOption[]>(() => {
    const missing = missingValues.value.map((value) => ({
        value,
        label: `${i18n.global.t('cronjob.websiteSelectionMissing')}: ${value}`,
    }));
    return missing.length ? [...availableOptions.value.nodes, ...missing] : availableOptions.value.nodes;
});

const changeSelection = (value: string[]) => {
    panelOrder.value = value || [];
    const next = new Set(value || []);
    if (next.has('all') && !props.modelValue.includes('all')) {
        emit('update:modelValue', ['all']);
        return;
    }
    if (next.size > 1) next.delete('all');
    for (const group of groups.value) {
        const token = `group:${group.id}`;
        const ids = members(group.id);
        if (!next.has(token)) {
            if (props.modelValue.includes(token)) ids.forEach((id) => next.delete(id));
            continue;
        }
        // Removing a member from a dynamic group turns the remaining members into a fixed list.
        if (props.modelValue.includes(token) && ids.some((id) => !next.has(id))) {
            next.delete(token);
        } else {
            ids.forEach((id) => next.delete(id));
        }
    }
    emit(
        'update:modelValue',
        [...next].filter((value) => !value.startsWith('menu:')),
    );
};
onMounted(async () => {
    try {
        const [sites, groupList] = await Promise.all([getWebsiteOptions({}), getAgentGroupList('website')]);
        websites.value = sites.data || [];
        groups.value = groupList.data || [];
    } finally {
        loading.value = false;
    }
});
</script>

<style>
/* Group rows only expand; the explicit All leaf selects the dynamic group. */
.el-cascader-node:has(.website-group-label) > .el-checkbox {
    display: none;
}
</style>
