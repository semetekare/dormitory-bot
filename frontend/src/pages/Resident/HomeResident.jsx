import { Panel, Grid, Container, Flex, Typography } from '@maxhub/max-ui';
import ButtonsPanelResident from '../../components/ButtonsPanelResident';
import Header from '../../components/Header';
import Footer from '../../components/Footer';
import WelcomeBlock from '../../components/WelcomeBlock';

const HomeResident = () => (
    <div style={{
        minHeight: '100vh',
    }}>
        <Header />
        <Container style={{ flex: 1 }}>
            <Panel mode="secondary" className="panel mt-3" style={{ height: '100%' }}>
                <Grid gap={12} cols={1}>
                    <WelcomeBlock />
                    <ButtonsPanelResident />
                </Grid>
            </Panel>
        </Container>
        <Footer />
    </div>
);

export default HomeResident;
