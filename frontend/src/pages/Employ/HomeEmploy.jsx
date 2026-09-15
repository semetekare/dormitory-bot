import { Panel, Grid, Container, Flex, Typography, ToolButton } from '@maxhub/max-ui';
import Header from '../../components/Header';
import Footer from '../../components/Footer';
import ButtonsPanelEmploy from '../../components/ButtonsPanelEmploy';
import WelcomeBlock from '../../components/WelcomeBlock';


const HomeEmploy = () => (
    <div style={{ minHeight: '100vh' }}>
        <Header />
        <Container style={{ flex: 1 }}>
            <Panel mode="secondary" className="panel mt-3" style={{ height: '100%' }}>
                <Grid gap={12} cols={1}>
                    <WelcomeBlock />
                    <ButtonsPanelEmploy />
                </Grid>
            </Panel>
        </Container>
        <Footer />
    </div>
);

export default HomeEmploy;
