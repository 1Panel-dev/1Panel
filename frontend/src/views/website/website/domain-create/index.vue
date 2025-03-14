<template>
    <div>
        <el-form-item :label="$t('website.batchAdd')">
            <el-row :gutter="20">
                <el-col :span="20">
                    <el-input
                        class="p-w-400"
                        type="textarea"
                        :rows="3"
                        v-model="create.domainStr"
                        :placeholder="$t('website.domainHelper')"
                    ></el-input>
                </el-col>
                <el-col :span="4">
                    <el-button @click="gengerateDomains" :disabled="create.domainStr == ''">
                        {{ $t('website.generateDomain') }}
                    </el-button>
                </el-col>
            </el-row>
        </el-form-item>
        <el-row :gutter="20" v-for="(domain, index) of create.domains" :key="index">
            <el-col :span="8">
                <el-form-item
                    :label="index == 0 ? $t('website.domain') : ''"
                    :prop="`domains.${index}.domain`"
                    :rules="rules.domain"
                >
                    <el-input
                        type="string"
                        v-model="create.domains[index].domain"
                        :placeholder="index > 0 ? $t('website.domain') : ''"
                    ></el-input>
                </el-form-item>
            </el-col>
            <el-col :span="8">
                <el-form-item
                    :label="index == 0 ? $t('commons.table.port') : ''"
                    :prop="`domains.${index}.port`"
                    :rules="rules.port"
                >
                    <el-input type="number" v-model.number="create.domains[index].port"></el-input>
                </el-form-item>
            </el-col>
            <el-col :span="4">
                <el-form-item :label="index == 0 ? 'SSL' : ''" prop="ssl">
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
                    <el-button @click="removeDomain(index)" link type="primary">
                        <el-icon><Delete /></el-icon>
                    </el-button>
                </el-form-item>
            </el-col>
        </el-row>
    </div>
</template>

<script lang="ts" setup>
import { Rules, checkNumberRange } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import { ref } from 'vue';

const emit = defineEmits(['gengerate']);
const props = defineProps({
    form: {
        type: Object,
        default: function () {
            return {};
        },
    },
});
const rules = ref({
    port: [Rules.requiredInput, Rules.paramPort, checkNumberRange(1, 65535)],
    domain: [Rules.requiredInput, Rules.domain],
    domains: {
        type: Array,
    },
});
const initDomain = () => ({
    domain: '',
    port: 80,
    ssl: false,
});
const create = ref({
    websiteID: 0,
    domains: [initDomain()],
    domainStr: '',
});

const addDomain = () => {
    create.value.domains.push(initDomain());
};

const removeDomain = (index: number) => {
    create.value.domains.splice(index, 1);
};

const checkDomain = (domain: string) => {
    const reg =
        /^([\w\u4e00-\u9fa5\-\*]{1,100}\.){1,10}([\w\u4e00-\u9fa5\-]{1,24}|[\w\u4e00-\u9fa5\-]{1,24}\.[\w\u4e00-\u9fa5\-]{1,24})(:\d{1,5})?$/;
    return reg.test(domain);
};

const gengerateDomains = () => {
    const lines = create.value.domainStr.split(/\r?\n/);
    lines.forEach((line) => {
        const [domain, port] = line.split(':');
        if (!checkDomain(domain)) {
            MsgError(line + i18n.global.t('commons.rule.domain'));
            return;
        }
        const exists = (domain: string, port: number): boolean => {
            return create.value.domains.some((info) => info.domain === domain && info.port === port);
        };
        if (exists(domain, port ? Number(port) : 80)) {
            return;
        }
        if (create.value.domains[0].domain == '') {
            create.value.domains[0].domain = domain;
            create.value.domains[0].port = port ? Number(port) : 80;
        } else {
            create.value.domains.push({
                domain,
                port: port ? Number(port) : 80,
                ssl: false,
            });
        }
    });
    emit('gengerate');
};

const handleParams = () => {
    props.form.domains = create.value.domains;
};

onMounted(() => {
    handleParams();
});
</script>
