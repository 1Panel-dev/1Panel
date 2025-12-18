<template>
    <div>
        <div class="d-description flex flex-wrap items-center justify-start gap-1.5">
            <el-button class="mr-1" plain size="small">
                {{ $t('app.version') }}{{ $t('commons.colon') }}{{ installed.version }}
            </el-button>
            <el-button v-if="installed.httpPort > 0" class="mr-1" plain size="small">
                {{ $t('commons.table.port') }}{{ $t('commons.colon') }}{{ installed.httpPort }}
            </el-button>
            <el-button v-if="installed.httpsPort > 0" plain size="small">
                {{ $t('commons.table.port') }}：{{ installed.httpsPort }}
            </el-button>

            <el-popover placement="right" trigger="hover" v-if="hasLinkButton(installed)" :width="400">
                <template #reference>
                    <el-button plain icon="Promotion" size="small" @click="$emit('openLink')">
                        {{ $t('app.toLink') }}
                    </el-button>
                </template>
                <table>
                    <tbody>
                        <tr v-if="defaultLink != ''">
                            <td v-if="installed.httpPort > 0">
                                <el-button
                                    type="primary"
                                    link
                                    @click="$emit('toLink', 'http://' + defaultLink + ':' + installed.httpPort)"
                                >
                                    {{ 'http://' + defaultLink + ':' + installed.httpPort }}
                                </el-button>
                            </td>
                        </tr>
                        <tr v-if="defaultLink != ''">
                            <td v-if="installed.httpsPort > 0">
                                <el-button
                                    type="primary"
                                    link
                                    @click="$emit('toLink', 'https://' + defaultLink + ':' + installed.httpsPort)"
                                >
                                    {{ 'https://' + defaultLink + ':' + installed.httpsPort }}
                                </el-button>
                            </td>
                        </tr>
                        <tr v-if="installed.webUI != ''">
                            <td>
                                <el-button type="primary" link @click="$emit('toLink')">
                                    {{ installed.webUI }}
                                </el-button>
                            </td>
                        </tr>
                    </tbody>
                </table>
                <span v-if="defaultLink == '' && installed.webUI == ''">
                    {{ $t('app.webUIConfig') }}
                    <el-link icon="Position" @click="$emit('jumpToPath')" type="primary">
                        {{ $t('firewall.quickJump') }}
                    </el-link>
                </span>
            </el-popover>
        </div>
        <div class="description">
            <span>
                {{ $t('app.alreadyRun') }}{{ $t('commons.colon') }}
                {{ getAge(installed.createdAt) }}
            </span>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { getAge } from '@/utils/util';

interface Props {
    installed: App.AppInstalled;
    defaultLink: string;
}
defineProps<Props>();

defineEmits(['openLink', 'toLink', 'jumpToPath']);

const hasLinkButton = (installed: any) => {
    return (
        (installed.appType == 'website' || installed.appKey?.startsWith('local')) &&
        (installed.httpPort > 0 || installed.httpsPort > 0 || installed.webUI != '')
    );
};
</script>

<style scoped lang="scss">
@use '@/views/app-store/index.scss';

.d-description {
    .el-button + .el-button {
        margin-left: 0;
    }
}
</style>
