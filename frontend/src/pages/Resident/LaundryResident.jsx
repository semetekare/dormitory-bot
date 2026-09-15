import { Flex, Grid, Panel } from "@maxhub/max-ui";
import Header from "../../components/Header";
import WashMachine from "../../components/WashMachine";
import Footer from "../../components/Footer";

const LaundryResident = () => (
    <div style={{ minHeight: '100vh' }}>
        <Header />
        <Panel centeredX centeredY className="mt-3">
            <Grid gap={12} cols={1}>
                <Flex gapX={12}>
                    <WashMachine number={1} status={'busy'}></WashMachine>
                    <WashMachine number={2} status={'free'}></WashMachine>
                    <WashMachine number={3} status={'my'}></WashMachine>
                </Flex>
                <Flex gapX={12}>
                    <WashMachine number={4} status={'my'}></WashMachine>
                    <WashMachine number={5} status={'busy'}></WashMachine>
                    <WashMachine number={6} status={'free'}></WashMachine>
                </Flex>
                <Flex gapX={12}>
                    <WashMachine number={7} status={'busy'}></WashMachine>
                    <WashMachine number={8} status={'my'}></WashMachine>
                    <WashMachine number={9} status={'free'}></WashMachine>
                </Flex>
            </Grid>
        </Panel>
        <Footer />
    </div>
);

export default LaundryResident;
