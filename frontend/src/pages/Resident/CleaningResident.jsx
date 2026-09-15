import Calendar from "../../components/Calendar";
import Footer from "../../components/Footer";
import Header from "../../components/Header";

const CleaningResident = () => (
    <>
        <div style={{
            minHeight: '100vh',
        }}>
            <Header />
            <Calendar></Calendar>
            <Footer />
        </div>
    </>
);

export default CleaningResident;