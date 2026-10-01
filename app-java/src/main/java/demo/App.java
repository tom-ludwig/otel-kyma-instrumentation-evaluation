package demo;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.concurrent.ThreadLocalRandom;

@SpringBootApplication
@RestController
public class App {

    private final JdbcTemplate jdbc;

    public App(JdbcTemplate jdbc) {
        this.jdbc = jdbc;
    }

    public static void main(String[] args) {
        SpringApplication.run(App.class, args);
    }

    @GetMapping("/rolldice")
    public String rollDice() throws InterruptedException {
        Thread.sleep(ThreadLocalRandom.current().nextInt(50));
        jdbc.queryForObject("SELECT 1", Integer.class);
        return "rolled: " + (ThreadLocalRandom.current().nextInt(6) + 1) + "\n";
    }

    @GetMapping("/healthz")
    public void healthz() {}
}
