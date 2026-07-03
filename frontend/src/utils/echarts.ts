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
    DataZoomComponent,
} from 'echarts/components';

import { CanvasRenderer, SVGRenderer } from 'echarts/renderers';

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
    DataZoomComponent,

    CanvasRenderer,
    SVGRenderer,
]);

export default echarts;
