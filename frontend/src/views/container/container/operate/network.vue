<template>
    <div>
        <ComplexTable v-if="tmpNetworks" style="width: 100%" :data="tmpNetworks">
            <el-table-column :label="$t('container.network')" min-width="200">
                <template #default="{ row }">
                    <el-select v-model="row.network" @change="changeNetwork(row)">
                        <el-option
                            v-for="(item, indexV) of networkOptions"
                            :key="indexV"
                            :value="item.option"
                            :label="item.option"
                        />
                    </el-select>
                </template>
            </el-table-column>
            <el-table-column :label="$t('container.macAddr')" min-width="200">
                <template #default="{ row }">
                    <el-input v-model="row.macAddr" />
                </template>
            </el-table-column>
            <el-table-column label="IPv4" min-width="200">
                <template #default="{ row }">
                    <el-input
                        v-model="row.ipv4"
                        :disabled="isIpDisabled(row)"
                        :placeholder="$t('container.inputIpv4')"
                    />
                </template>
            </el-table-column>
            <el-table-column label="IPv6" min-width="200">
                <template #default="{ row }">
                    <el-input
                        v-model="row.ipv6"
                        :disabled="isIpDisabled(row)"
                        :placeholder="$t('container.inputIpv6')"
                    />
                </template>
            </el-table-column>
            <el-table-column min-width="80" :fixed="'right'">
                <template #default="scope">
                    <el-button link type="primary" @click="handleNetworksDelete(scope.$index)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
            </el-table-column>
        </ComplexTable>
        <el-button class="mt-2" @click="handleNetworksAdd()">
            {{ $t('commons.button.add') }}
        </el-button>
    </div>
</template>

<script lang="ts" setup>
import { Container } from '@/api/interface/container';
import { listNetwork } from '@/api/modules/container';
import { ref, onMounted } from 'vue';
const tmpNetworks = ref([]);
const networkOptions = ref();
const props = defineProps({
    networks: { type: Array<Container.ContainerNetwork>, default: [] },
});
watch(
    () => props.networks,
    (newVal) => {
        tmpNetworks.value = newVal || [];
        for (const item of tmpNetworks.value) {
            if (item.network === 'none' || item.network === 'host') {
                item.ipv4 = '';
                item.ipv6 = '';
            }
        }
    },
);

const loadNetworkOptions = async () => {
    const res = await listNetwork();
    networkOptions.value = res.data;
};
const handleNetworksAdd = () => {
    let item = {
        network: 'bridge',
        ipv4: '',
        ipv6: '',
        macAddr: '',
    };
    tmpNetworks.value.push(item);
};
const handleNetworksDelete = (index: number) => {
    tmpNetworks.value.splice(index, 1);
};
const changeNetwork = (row: any) => {
    if (row.network === 'none' || row.network === 'host') {
        row.ipv4 = '';
        row.ipv6 = '';
    }
};
const isIpDisabled = (row: any) => {
    return row.network === 'none' || row.network === 'host' || row.network === 'bridge';
};

const loadNetworks = () => {
    return tmpNetworks.value;
};

onMounted(() => {
    tmpNetworks.value = props.networks || [];
    loadNetworkOptions();
});

defineExpose({
    loadNetworks,
});
</script>
