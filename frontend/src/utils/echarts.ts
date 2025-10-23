import * as echarts from 'echarts/core';

import { LineChart, BarChart, PieChart, MapChart } from 'echarts/charts';

import {
    TitleComponent,
    TooltipComponent,
    GridComponent,
    LegendComponent,
    VisualMapComponent,
    GeoComponent,
    PolarComponent,
} from 'echarts/components';

import { CanvasRenderer } from 'echarts/renderers';

echarts.use([
    LineChart,
    BarChart,
    PieChart,
    MapChart,

    TitleComponent,
    TooltipComponent,
    GridComponent,
    LegendComponent,
    VisualMapComponent,
    GeoComponent,
    PolarComponent,

    CanvasRenderer,
]);

export default echarts;
