<template>
    <div>
        <el-descriptions :column="1" :label-width="isMobile ? 130 : 150" border size="small" class="break-all">
            <el-descriptions-item :label="$t('commons.table.name')">
                {{ item.kind === 'legacy' ? $t('apiKeyManagement.legacy') : item.name }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('apiKeyManagement.identifier')">{{ item.keyHint }}</el-descriptions-item>
            <el-descriptions-item :label="$t('commons.table.status')">
                <el-tag :type="item.status === 'Enable' ? 'success' : item.status === 'Disable' ? 'warning' : 'info'">
                    {{ $t('apiKeyManagement.status' + item.status) }}
                </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('setting.ipWhiteList')">
                <span class="whitespace-pre-wrap">
                    {{
                        isAnyAPIKeyIP(item.ipWhiteList)
                            ? $t('apiKeyManagement.anyIP')
                            : item.ipWhiteList.replace(/,/g, '\n') || '—'
                    }}
                </span>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('setting.apiTrustedProxies')">
                <span class="whitespace-pre-wrap">{{ item.apiTrustedProxies.replace(/,/g, '\n') || '—' }}</span>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('apiKeyManagement.expiresAt')">
                {{ item.expiresAt ? new Date(item.expiresAt).toLocaleString() : $t('apiKeyManagement.never') }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('apiKeyManagement.signatureWindow')">
                {{ item.apiKeyValidityTime }} {{ $t('commons.units.minute') }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('apiKeyManagement.allowAppBinding')">
                {{ $t(item.allowAppBinding ? 'commons.true' : 'commons.false') }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('commons.table.description')">
                {{ item.description || '—' }}
            </el-descriptions-item>
        </el-descriptions>
    </div>
</template>

<script setup lang="ts">
import type { APIKey } from '@/api/interface/api-key';
import { isAnyAPIKeyIP } from '@/utils/api-key';
import { useGlobalStore } from '@/composables/useGlobalStore';
defineProps<{ item: APIKey.Item }>();
const { isMobile } = useGlobalStore();
</script>
