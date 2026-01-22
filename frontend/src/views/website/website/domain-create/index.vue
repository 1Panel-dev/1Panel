<template>
    <div>
        <el-row :gutter="22" v-for="(domain, index) of create.domains" :key="index">
            <el-col :span="6">
                <el-form-item
                    :label="index == 0 ? $t('website.domain') : ''"
                    :prop="`domains.${index}.domain`"
                    :rules="rules.domain"
                >
                    <el-input
                        type="string"
                        v-model="create.domains[index].domain"
                        :placeholder="index > 0 ? $t('website.domain') : ''"
                        @blur="handleDomainBlur(index)"
                    ></el-input>
                    <span class="input-help" v-if="domainWarnings[index]">{{ $t('website.domainNotFQDN') }}</span>
                </el-form-item>
            </el-col>
            <el-col :span="6">
                <el-form-item :label="index == 0 ? $t('toolbox.device.hostname') : ''">
                    <el-input
                        type="string"
                        :model-value="create.domains[index].host"
                        :placeholder="index > 0 ? $t('toolbox.device.hostname') : ''"
                        disabled
                    ></el-input>
                </el-form-item>
            </el-col>
            <el-col :span="4">
                <el-form-item
                    :label="index == 0 ? $t('commons.table.port') : ''"
                    :prop="`domains.${index}.port`"
                    :rules="rules.port"
                >
                    <el-input type="number" v-model.number="create.domains[index].port"></el-input>
                </el-form-item>
            </el-col>
            <el-col :span="2">
                <el-form-item :label="index == 0 ? 'SSL' : ''" :prop="`domains.${index}.ssl`">
                    <el-checkbox
                        v-model="create.domains[index].ssl"
                        :disabled="create.domains[index].port == 80"
                    ></el-checkbox>
                </el-form-item>
            </el-col>
            <el-col :span="4" v-if="index == 0">
                <el-form-item :label="$t('commons.button.add') + $t('website.domain')">
                    <el-button @click="addDomain">
                        <el-icon><Plus /></el-icon>
                    </el-button>
                </el-form-item>
            </el-col>
            <el-col :span="4" v-else>
                <el-form-item>
                    <el-button @click="removeDomain(index)">
                        <el-icon><Delete /></el-icon>
                    </el-button>
                </el-form-item>
            </el-col>
        </el-row>
        <el-button @click="openBatchDialog" type="primary" plain>
            {{ $t('website.batchInput') }}
        </el-button>

        <el-dialog v-model="batchDialogVisible" :title="$t('website.batchAdd')" width="600px">
            <el-input
                type="textarea"
                :rows="8"
                v-model="create.domainStr"
                :placeholder="$t('website.domainBatchHelper')"
            ></el-input>
            <template #footer>
                <el-button @click="batchDialogVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="saveBatchInput" :disabled="create.domainStr == ''">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import { checkAppInstalled } from '@/api/modules/app';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { MsgError } from '@/utils/message';
import { ref, onMounted, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';

const i18n = useI18n();

const props = defineProps({
    form: {
        type: Object,
        default: function () {
            return {};
        },
    },
});

const emit = defineEmits(['gengerate']);

const ipv4Regex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;

const ipv6Regex =
    /^(?:(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,7}:|(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}|(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}|(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}|(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:(?:(?::[0-9a-fA-F]{1,4}){1,6})|:(?:(?::[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(?::[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(?:ffff(?::0{1,4}){0,1}:){0,1}(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])|(?:[0-9a-fA-F]{1,4}:){1,4}:(?:(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(?:25[0-5]|(?:2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$/;

const singleDomainRegex = /^(?:\*|[\w\u4e00-\u9fa5-]{1,63})(?:\.(?:\*|[\w\u4e00-\u9fa5-]{1,63}))*$/;
const FQDNDomainRegex =
    /^(?:\*\.)?(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}(?::(6553[0-5]|655[0-2]\d|65[0-4]\d{2}|6[0-4]\d{3}|[1-5]?\d{1,4}))?$/;

const isIPv4 = (str: string): boolean => {
    return ipv4Regex.test(str);
};

const isIPv6 = (str: string): boolean => {
    const cleanStr = str.replace(/^\[|\]$/g, '');
    return ipv6Regex.test(cleanStr);
};

const isFQDN = (domain: string): boolean => {
    if (!domain) return true;
    return FQDNDomainRegex.test(domain);
};

function toPunycode(hostname: string): string {
    try {
        if (isIPv4(hostname) || isIPv6(hostname.replace(/^\[|\]$/g, ''))) {
            return hostname;
        }

        if (hostname.startsWith('*.')) {
            const domainPart = hostname.substring(2);
            const convertedDomain = new URL(`http://${domainPart}`).hostname;
            return `*.${convertedDomain}`;
        }
        if (hostname.endsWith('.*')) {
            const domainPart = hostname.slice(0, -2);
            const convertedDomain = new URL(`http://${domainPart}`).hostname;
            return `${convertedDomain}.*`;
        }
        return new URL(`http://${hostname}`).hostname;
    } catch {
        return hostname;
    }
}

const handleDomainBlur = (index: number) => {
    const originalDomain = create.value.domains[index].domain;
    if (!originalDomain) {
        create.value.domains[index].host = '';
        domainWarnings.value[index] = false;
        return;
    }

    const cleanDomain = originalDomain.replace(/^\[|\]$/g, '');
    if (isIPv4(cleanDomain) || isIPv6(cleanDomain)) {
        create.value.domains[index].host = originalDomain;
        create.value.domains[index].domain = originalDomain;
        domainWarnings.value[index] = false;
        return;
    }

    let singleRegexChecked = singleDomainRegex.test(originalDomain);
    if (!singleRegexChecked) {
        domainWarnings.value[index] = false;
        return;
    }

    const punycoded = toPunycode(originalDomain);
    if (punycoded !== originalDomain) {
        create.value.domains[index].host = originalDomain;
        create.value.domains[index].domain = punycoded;
    } else {
        create.value.domains[index].host = originalDomain;
    }

    if (singleRegexChecked) {
        domainWarnings.value[index] = !isFQDN(create.value.domains[index].domain);
    } else {
        domainWarnings.value[index] = false;
    }
};

const validateSingleDomain = (_rule: any, value: string, callback: (error?: Error) => void) => {
    if (!value) {
        return callback();
    }

    const cleanValue = value.replace(/^\[|\]$/g, '');

    if (isIPv4(cleanValue) || isIPv6(cleanValue) || singleDomainRegex.test(value)) {
        return callback();
    }

    callback(new Error(i18n.t('website.domainInvalid')));
};

const rules = ref({
    port: [Rules.requiredInput, Rules.paramPort, checkNumberRange(1, 65535)],
    domain: [Rules.requiredInput, { validator: validateSingleDomain, trigger: 'blur' }],
    domains: {
        type: Array,
    },
});
const defaultPort = ref(80);
const batchDialogVisible = ref(false);
const domainWarnings = ref<{ [key: number]: boolean }>({});

const initDomain = () => ({
    domain: '',
    host: '',
    port: defaultPort.value,
    ssl: false,
});
const create = ref({
    websiteID: 0,
    domains: [initDomain()],
    domainStr: '',
});

const domainToString = (domain: { domain: string; port: number; ssl: boolean }): string => {
    if (!domain.domain) return '';
    let str = domain.domain;

    if (domain.port && domain.port !== 80 && isIPv6(str.replace(/^\[|\]$/g, ''))) {
        if (!str.startsWith('[')) {
            str = `[${str}]`;
        }
    }

    if (domain.port && domain.port !== 80) {
        str += `:${domain.port}`;
    }
    return str;
};

const stringToDomain = (line: string): { domain: string; host: string; port: number; ssl: boolean } | null => {
    if (!line.trim()) return null;

    let ssl = false;
    let str = line.trim();
    let domainRaw = '';
    let portStr = '';

    const ipv6Match = str.match(/^\[([^\]]+)\](?::(\d+))?$/);
    if (ipv6Match) {
        domainRaw = ipv6Match[1];
        portStr = ipv6Match[2] || '';
    } else {
        const parts = str.split(':');
        if (parts.length === 2) {
            domainRaw = parts[0];
            portStr = parts[1];
        } else if (parts.length === 1) {
            domainRaw = parts[0];
        } else {
            if (isIPv6(str)) {
                domainRaw = str;
            } else {
                return null;
            }
        }
    }

    if (!domainRaw) return null;

    const cleanDomain = domainRaw.replace(/^\[|\]$/g, '');
    let domain: string;
    let host: string;

    if (isIPv4(cleanDomain) || isIPv6(cleanDomain)) {
        domain = domainRaw;
        host = domainRaw;
    } else {
        domain = toPunycode(domainRaw);
        host = domain !== domainRaw ? domainRaw : domain;
    }

    const port = portStr ? Number(portStr) : defaultPort.value;

    if (port === 443) {
        ssl = true;
    } else if (port === 80) {
        ssl = false;
    }

    return { domain, host, port, ssl };
};

const openBatchDialog = () => {
    const lines = create.value.domains.map(domainToString).filter((line) => line !== '');
    create.value.domainStr = lines.join('\n');
    batchDialogVisible.value = true;
};

const addDomain = () => {
    create.value.domains.push(initDomain());
};

const removeDomain = (index: number) => {
    create.value.domains.splice(index, 1);
};

const saveBatchInput = () => {
    if (create.value.domainStr.trim() === '') {
        return true;
    }
    const lines = create.value.domainStr.split(/\r?\n/);
    const newDomains: { domain: string; host: string; port: number; ssl: boolean }[] = [];
    let hasError = false;

    const isFirstBatchInput = create.value.domains.length === 1 && !create.value.domains[0].domain;

    lines.forEach((line, index) => {
        const parsed = stringToDomain(line);
        if (!parsed) return;

        const { domain, host, port, ssl } = parsed;
        const cleanDomain = domain.replace(/^\[|\]$/g, '');

        if (!domain.trim() || (!isIPv4(cleanDomain) && !isIPv6(cleanDomain) && !singleDomainRegex.test(domain))) {
            MsgError(line + ' ' + i18n.t('website.domainInvalid'));
            hasError = true;
            return;
        }

        if (!newDomains.some((d) => d.domain === domain && d.port === port)) {
            newDomains.push({ domain, host, port, ssl });
        }

        if (isFirstBatchInput && index === 0) {
            const alias = domain.split(':')[0].replace(/^\[|\]$/g, '');
            if (props.form.alias !== undefined) {
                props.form.alias = alias;
            }
        }
    });

    if (hasError) return false;

    if (newDomains.length > 0) {
        create.value.domains = newDomains;
    } else {
        create.value.domains = [initDomain()];
    }

    handleParams();
    nextTick(() => emit('gengerate'));

    batchDialogVisible.value = false;
    return true;
};

const handleParams = () => {
    props.form.domains = create.value.domains;
};

const getOprensty = async () => {
    try {
        await checkAppInstalled('openresty', '')
            .then((res) => {
                defaultPort.value = res.data.httpPort || 80;
                create.value.domains[0].port = defaultPort.value;
            })
            .catch(() => {});
    } catch (error) {}
};

onMounted(() => {
    getOprensty();
    handleParams();
});
</script>
